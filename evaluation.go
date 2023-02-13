package featbit

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
)

const (
	ExptKeyPrefix          = "expt"
	ReasonFlagOff          = "flag off"
	ReasonTargetMatch      = "target match"
	ReasonRuleMatch        = "rule match"
	ReasonFallthrough      = "fall through all rules"
	ReasonClientNotReady   = "client not ready"
	ReasonFlagNotFound     = "flag not found"
	ReasonWrongType        = "wrong type"
	ReasonUserNotSpecified = "user not specified"
	ReasonError            = "error in evaluation"
	FlagKeyUnknown         = "flag key unknown"
	FlagNameUnknown        = "flag Name unknown"
	FlagValueUnknown       = "flag value unknown"
	ThanClause             = "Than"
	GeClause               = "BiggerEqualThan"
	GtClause               = "BiggerThan"
	LeClause               = "LessEqualThan"
	LtClause               = "LessThan"
	EqClause               = "Equal"
	NeqClause              = "NotEqual"
	ContainsClause         = "Contains"
	NotContainClause       = "NotContain"
	IsOneOfClause          = "IsOneOf"
	NotOneOfClause         = "NotOneOf"
	StartsWithClause       = "StartsWith"
	EndsWithClause         = "EndsWith"
	IsTrueClause           = "IsTrue"
	IsFalseClause          = "IsFalse"
	MatchRegexClause       = "MatchRegex"
	NotMatchRegexClause    = "NotMatchRegex"
	IsInSegmentClause      = "User is in segment"
	NotInSegmentClause     = "User is not in segment"
	FlagJsonType           = "json"
	FlagBoolType           = "boolean"
	FlagNumericType        = "number"
	FlagStringType         = "string"
)

type evaluator struct {
	getFlag    func(key string) *data.FeatureFlag
	getSegment func(key string) *data.Segment
	funcSlice  []func(*data.FeatureFlag, *FBUser) (evalResult, bool)
}

func newEvaluator(getFlag func(key string) *data.FeatureFlag,
	getSegment func(key string) *data.Segment) *evaluator {
	e := &evaluator{getFlag: getFlag, getSegment: getSegment}
	fs := []func(*data.FeatureFlag, *FBUser) (evalResult, bool){
		e.matchFeatureFlagDisabledUserVariation,
		e.matchTargetedUserVariation,
		e.matchConditionedUserVariation,
		e.matchFallThroughUserVariation,
	}
	e.funcSlice = fs
	return e
}

func (e *evaluator) evaluate(flag *data.FeatureFlag, user *FBUser, event Event) (er evalResult) {
	defer func() {
		if er.success {
			log.LogInfo("FB Go SDK: User %v, Feature Flag %v, Flag Value %v", user.GetKey(), flag.Key, er.fv)
			if event != nil {
				eventFlag := er.toEventFlag()
				event.Add(eventFlag)
			}
		}
	}()
	var ok bool
	for _, f := range e.funcSlice {
		er, ok = f(flag, user)
		if ok {
			return
		}
	}
	return
}

func (e *evaluator) matchFeatureFlagDisabledUserVariation(flag *data.FeatureFlag, _ *FBUser) (evalResult, bool) {
	if !flag.Enabled {
		return evalResult{
			id: flag.DisabledVariationId,
			detail: detail{
				Reason:  ReasonFlagOff,
				KeyName: flag.Key,
				Name:    flag.Name,
			},
			sendToExperiment: false,
			fv:               flag.GetFlagValue(flag.DisabledVariationId),
			success:          true,
			flagType:         flag.VariationType,
		}, true
	}
	return evalResult{}, false
}

func (e *evaluator) matchTargetedUserVariation(flag *data.FeatureFlag, user *FBUser) (evalResult, bool) {
	for _, targetUser := range flag.TargetUsers {
		for _, keyId := range targetUser.KeyIds {
			if keyId == user.GetKey() {
				return evalResult{
					id: targetUser.VariationId,
					detail: detail{
						Reason:  ReasonTargetMatch,
						KeyName: flag.Key,
						Name:    flag.Name,
					},
					sendToExperiment: flag.ExptIncludeAllTargets,
					fv:               flag.GetFlagValue(targetUser.VariationId),
					success:          true,
					flagType:         flag.VariationType,
				}, true
			}
		}
	}
	return evalResult{}, false
}

func (e *evaluator) matchConditionedUserVariation(flag *data.FeatureFlag, user *FBUser) (evalResult, bool) {
	var rule *data.TargetRule
	for _, targetRule := range flag.Rules {
		if e.ifUserMatchRule(user, targetRule.Conditions) {
			rule = &targetRule
			break
		}
	}
	if rule != nil {
		return getRolloutVariationValue(flag, rule.Variations, user, ReasonRuleMatch, rule.IncludedInExpt, rule.DispatchKey)
	}
	return evalResult{}, false
}

func (e *evaluator) matchFallThroughUserVariation(flag *data.FeatureFlag, user *FBUser) (evalResult, bool) {
	ft := flag.Fallthrough
	return getRolloutVariationValue(flag, ft.Variations, user, ReasonFallthrough, ft.IncludedInExpt, ft.DispatchKey)
}
