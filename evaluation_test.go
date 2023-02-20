package featbit

import (
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/datastorage"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/stretchr/testify/assert"
	"testing"
)

var user1, _ = interfaces.NewUserBuilder("test-user-1").Build()

var user2, _ = interfaces.NewUserBuilder("test-target-user").Build()

var user3, _ = interfaces.NewUserBuilder("test-true-user").Custom("graduated", "true").Build()

var user4, _ = interfaces.NewUserBuilder("test-equal-user").Custom("country", "CHN").Build()

var user5, _ = interfaces.NewUserBuilder("test-than-user").Custom("salary", "2500").Build()

var user6, _ = interfaces.NewUserBuilder("test-contain-user").Custom("email", "test-contain-user@gmail.com").Build()

var user7, _ = interfaces.NewUserBuilder("test-isoneof-user").Custom("major", "CS").Build()

var user8, _ = interfaces.NewUserBuilder("group-admin-user").Build()

var user9, _ = interfaces.NewUserBuilder("test-regex-user").Custom("phone", "18555358000").Build()

var user10, _ = interfaces.NewUserBuilder("test-fallthrough-user").Build()

var disabledFlag *data.FeatureFlag

var flag *data.FeatureFlag

var eval *evaluator

func init() {
	dataStorage, _ := datastorage.NewMockDataStorageBuilder().CreateDataStorage(nil)
	mockDataStorage := dataStorage.(*datastorage.MockDataStorage)
	_ = mockDataStorage.LoadData()

	getFlag := func(key string) *data.FeatureFlag {
		if item, e := dataStorage.Get(data.Features, key); e == nil {
			if flag, ok := item.(*data.FeatureFlag); ok {
				return flag
			}
		}
		return nil
	}
	getSegment := func(key string) *data.Segment {
		if item, e := dataStorage.Get(data.Segments, key); e == nil {
			if segment, ok := item.(*data.Segment); ok {
				return segment
			}
		}
		return nil
	}
	eval = newEvaluator(getFlag, getSegment)
	disabledFlag = getFlag("ff-test-off")
	flag = getFlag("ff-evaluation-test")
}

func TestEvaluation(t *testing.T) {
	t.Run("evaluation for disable flag", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user1))
		er := eval.evaluate(disabledFlag, &user1, event)
		assert.Equal(t, "false", er.fv)
		assert.Equal(t, ReasonFlagOff, er.reason)
	})
	t.Run("evaluation for target user", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user2))
		er := eval.evaluate(flag, &user2, event)
		assert.Equal(t, "teamB", er.fv)
		assert.Equal(t, ReasonTargetMatch, er.reason)
	})
	t.Run("evaluation for true condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user3))
		er := eval.evaluate(flag, &user3, event)
		assert.Equal(t, "teamC", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for equal condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user4))
		er := eval.evaluate(flag, &user4, event)
		assert.Equal(t, "teamD", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for than condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user5))
		er := eval.evaluate(flag, &user5, event)
		assert.Equal(t, "teamE", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for contain condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user6))
		er := eval.evaluate(flag, &user6, event)
		assert.Equal(t, "teamF", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for one of condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user7))
		er := eval.evaluate(flag, &user7, event)
		assert.Equal(t, "teamG", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for start end condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user8))
		er := eval.evaluate(flag, &user8, event)
		assert.Equal(t, "teamH", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for regex condition", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user9))
		er := eval.evaluate(flag, &user9, event)
		assert.Equal(t, "teamI", er.fv)
		assert.Equal(t, ReasonRuleMatch, er.reason)
	})
	t.Run("evaluation for fall through", func(t *testing.T) {
		event := insight.NewFlagEvent(insight.ConvertFBUserToEventUser(&user10))
		er := eval.evaluate(flag, &user10, event)
		assert.Equal(t, "teamA", er.fv)
		assert.Equal(t, ReasonFallthrough, er.reason)
	})
}
