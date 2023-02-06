package datasynchronization

import (
	realRand "crypto/rand"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"math"
	"math/big"
	"math/rand"
	"time"
)

type BackoffAndJitterStrategy struct {
	firstRetryDelay time.Duration
	maxRetryDelay   time.Duration
	resetInterval   time.Duration
	jitterRatio     float64
	retryCount      float64
	lastGoodRun     time.Time
}

var defaultStrategy = &BackoffAndJitterStrategy{
	maxRetryDelay: 60 * 1000 * time.Millisecond,
	resetInterval: 60 * 1000 * time.Millisecond,
	jitterRatio:   0.5,
	retryCount:    0,
}

func NewWithFirstRetryDelay(firstRetryDelay time.Duration) *BackoffAndJitterStrategy {
	defaultStrategy.firstRetryDelay = firstRetryDelay
	return defaultStrategy
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
	log.LogInfo("backoff: %v, jitter: %v, next delay: %v", backOff, jitterTime, millis.Milliseconds())
	return millis
}
