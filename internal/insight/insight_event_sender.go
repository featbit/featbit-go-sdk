package insight

import (
	"bytes"
	"context"
	"fmt"
	"github.com/featbit/featbit-go-sdk/v2/interfaces"
	"github.com/featbit/featbit-go-sdk/v2/internal/util/log"
	"io/ioutil"
	"log/slog"
	"net/http"
	"time"
)

var invalidInput = fmt.Errorf("invalid url or json")

type EventSenderImp struct {
	logger        interfaces.Logger
	client        *http.Client
	headers       http.Header
	retryInterval time.Duration
	maxRetryTimes int
}

func NewEventSenderImp(client *http.Client, headers http.Header, retryInterval time.Duration, maxRetryTimes int, loggers ...interfaces.Logger) *EventSenderImp {
	logger := log.FromContext(nil)
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}
	return &EventSenderImp{logger: logger, client: client, headers: headers, retryInterval: retryInterval, maxRetryTimes: maxRetryTimes}
}

func (e *EventSenderImp) PostJson(uri string, jsonBytes []byte) ([]byte, error) {

	if uri == "" || len(jsonBytes) == 0 {
		return nil, invalidInput
	}

	headers := make(http.Header)
	for k, vv := range e.headers {
		headers[k] = vv
	}
	headers.Set("Content-Type", "application/json")

	var resp *http.Response
	var respErr error
	for attempt := 0; attempt <= e.maxRetryTimes; attempt++ {
		if attempt > 0 {
			delay := e.retryInterval << attempt
			if delay > time.Second {
				delay = time.Second
			}
			time.Sleep(delay)
		}

		req, reqErr := http.NewRequest("POST", uri, bytes.NewReader(jsonBytes))
		if reqErr != nil {
			e.logger.Log(context.Background(), slog.LevelError, "FB GO SDK: events sending error", "error", reqErr)
			return nil, reqErr
		}
		req.Header = headers
		resp, respErr = e.client.Do(req)
		// close http client， ignore response body
		if resp != nil && resp.Body != nil {
			_, _ = ioutil.ReadAll(resp.Body)
			_ = resp.Body.Close()
		}
		if respErr != nil {
			e.logger.Log(context.Background(), slog.LevelError, "FB GO SDK: events sending error", "error", respErr)
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			e.logger.Log(context.Background(), slog.LevelDebug, "sending events ok")
			return nil, nil
		}
	}
	return nil, respErr

}

func (e *EventSenderImp) Close() error {
	return nil
}
