package insight

import (
	"encoding/json"
	"fmt"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal"
	"github.com/featbit/featbit-go-sdk/internal/mocks"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

var ctx, _ = internal.FromConfig("fake env scret", "fake url", "fake url", nil)

var user1, _ = NewUserBuilder("test-user-1").Build()

var user2, _ = NewUserBuilder("test-user-2").Build()

var user3, _ = NewUserBuilder("test-user-3").Build()

var f = func(bytes []byte) []Event {
	var events []*insight.UserEvent
	_ = json.Unmarshal(bytes, &events)
	ret := make([]Event, len(events))
	for i, event := range events {
		ret[i] = event
	}
	return ret
}

func TestInsightProcessor(t *testing.T) {
	t.Run("start and close", func(t *testing.T) {
		sender := mocks.NewMockSender()
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		_, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.False(t, ok)
		_ = insightProcessor.Close()
	})
	t.Run("start and close if sender error on close", func(t *testing.T) {
		sender := mocks.NewMockSender()
		sender.SetCloseErr(fmt.Errorf("fake error"))
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		_, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.False(t, ok)
		_ = insightProcessor.Close()
	})
	t.Run("send events and auto flush", func(t *testing.T) {
		sender := mocks.NewMockSender()
		sender.SetParseJson(f)
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user1)))
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user2)))
		res, _ := sender.GetLatestSendingInfo(200 * time.Millisecond)
		if res.Size() == 1 {
			assert.True(t, res.Contains("test-user-1"))
		} else {
			assert.True(t, res.Contains("test-user-1"))
			assert.True(t, res.Contains("test-user-2"))
		}
		_ = insightProcessor.Close()
	})
	t.Run("send events and manuel flush", func(t *testing.T) {
		sender := mocks.NewMockSender()
		sender.SetParseJson(f)
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user1)))
		insightProcessor.Flush()
		time.Sleep(10 * time.Millisecond)
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user2)))
		insightProcessor.Flush()
		res, _ := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 1, res.Size())
		assert.True(t, res.Contains("test-user-1"))
		res, _ = sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 1, res.Size())
		assert.True(t, res.Contains("test-user-2"))
		_ = insightProcessor.Close()

	})
	t.Run("still work even if error in sending", func(t *testing.T) {
		sender := mocks.NewMockSender()
		sender.SetParseJson(f)
		sender.SetErr(fmt.Errorf("fake err"))
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user1)))
		insightProcessor.Flush()
		time.Sleep(10 * time.Millisecond)
		_, _ = sender.GetLatestSendingInfo(200 * time.Millisecond)
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user2)))
		insightProcessor.Flush()
		res, _ := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 1, res.Size())
		assert.True(t, res.Contains("test-user-2"))
		_ = insightProcessor.Close()
	})
	t.Run("can't send anything after close", func(t *testing.T) {
		sender := mocks.NewMockSender()
		sender.SetParseJson(f)
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		_ = insightProcessor.Close()
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user1)))
		insightProcessor.Flush()
		_, ok := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.False(t, ok)
	})
	t.Run("event keeping in buffer if all flush workers are busy", func(t *testing.T) {
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(5)
		sender := mocks.NewMockSender()
		sender.SetWaitGroup(waitGroup)
		sender.SetParseJson(f)
		sender.MustWait()
		insightProcessor := NewEventProcessor(ctx, sender, 100, 100*time.Millisecond)
		for i := 0; i < 5; i++ {
			insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user1)))
			insightProcessor.Flush()
			_, _ = sender.GetLatestSendingInfo(200 * time.Millisecond)
		}
		waitGroup.Wait()
		sender.NoMoreWait()
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user2)))
		insightProcessor.Flush()
		insightProcessor.Send(insight.NewUserEvent(insight.ConvertFBUserToEventUser(&user3)))
		insightProcessor.Flush()
		time.Sleep(10 * time.Millisecond)
		sender.Completed()
		res, _ := sender.GetLatestSendingInfo(200 * time.Millisecond)
		assert.Equal(t, 2, res.Size())
		assert.True(t, res.Contains("test-user-2"))
		assert.True(t, res.Contains("test-user-3"))
		_ = insightProcessor.Close()
	})

}
