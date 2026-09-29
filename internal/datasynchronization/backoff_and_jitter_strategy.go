package datasynchronization

import (
	realRand "crypto/rand"
	"github.com/featbit/featbit-go-sdk/v2/internal/util/log"
	"log/slog"
	"math"
	"math/big"
	"math/rand"
	"time"
)

type BackoffAndJitterStrategy struct {
	logger          *slog.Logger
	firstRetryDelay time.Duration
	maxRetryDelay   time.Duration
	resetInterval   time.Duration
	jitterRatio     float64
	retryCount      float64
	lastGoodRun     time.Time
}

func NewWithFirstRetryDelay(firstRetryDelay time.Duration, loggers ...*slog.Logger) *BackoffAndJitterStrategy {
	logger := log.FromContext(nil)
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}
	return &BackoffAndJitterStrategy{
		logger:          logger,
		firstRetryDelay: firstRetryDelay,
		maxRetryDelay:   60 * time.Second,
		resetInterval:   60 * time.Second,
		jitterRatio:     0.5,
	}
}

func (s *BackoffAndJitterStrategy) SetGoodRunAtNow() {
	s.lastGoodRun = time.Now()
}

func (s *BackoffAndJitterStrategy) countBackoffTime() float64 {
	delay := s.firstRetryDelay.Seconds() * math.Pow(2, s.retryCount)
	return math.Min(s.maxRetryDelay.Seconds(), delay)
}

func (s *BackoffAndJitterStrategy) countJitterTime(delay float64) float64 {
	rv, err := realRand.Int(realRand.Reader, big.NewInt(100))
	if err != nil {
		rv = big.NewInt(rand.Int63n(100))
	}
	return delay * s.jitterRatio * float64(rv.Int64()) / 100
}

func (b *BackoffAndJitterStrategy) NextDelay() time.Duration {
	now := time.Now()
	interval := now.Sub(b.lastGoodRun)
	if interval > b.resetInterval {
		b.retryCount = 0
	}
	backOff := b.countBackoffTime()
	jitterTime := b.countJitterTime(backOff)
	delay := (jitterTime + backOff/2) * 1000
	b.retryCount += 1
	millis := time.Duration(int64(math.Floor(delay))) * time.Millisecond
	b.logger.Info("backoff before retry", "backoff_seconds", backOff, "jitter_seconds", jitterTime, "delay", millis)
	return millis
}
