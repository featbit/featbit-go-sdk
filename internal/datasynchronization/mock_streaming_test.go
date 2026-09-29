package datasynchronization

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestMockStreamingConcurrentInitialization(t *testing.T) {
	for _, tc := range []struct {
		name    string
		success bool
	}{
		{name: "success", success: true},
		{name: "failure", success: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := &MockStreaming{success: tc.success}
			if stream.IsInitialized() {
				t.Fatal("stream should not be initialized before Start")
			}
			stop := make(chan struct{})
			var readersReady, readersDone sync.WaitGroup
			defer func() {
				close(stop)
				readersDone.Wait()
			}()
			for i := 0; i < 8; i++ {
				readersReady.Add(1)
				readersDone.Add(1)
				go func() {
					defer readersDone.Done()
					readersReady.Done()
					for !stream.IsInitialized() {
						select {
						case <-stop:
							return
						default:
							runtime.Gosched()
						}
					}
				}()
			}
			// Poll while Start's goroutine updates initialization state. Waiting
			// for Start to finish before launching readers would hide the race.
			readersReady.Wait()
			select {
			case <-stream.Start():
			case <-time.After(3 * time.Second):
				t.Fatal("stream initialization did not finish")
			}
			if got := stream.IsInitialized(); got != tc.success {
				t.Errorf("IsInitialized() = %v, want %v", got, tc.success)
			}
		})
	}
}
