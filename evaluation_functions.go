package featbit

import (
	"encoding/json"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/data"
	"github.com/featbit/featbit-go-sdk/internal/util"
	"regexp"
	"strconv"
	"strings"
)

func (e *evaluator) ifUserMatchRule(user *FBUser, conditions []data.Condition) bool {
	for _, condition := range conditions {
		if e.ifUserMatchCondition(user, &condition) {
			continue
		}
		return false
	}
	return true
}

func (e *evaluator) ifUserMatchCondition(user *FBUser, condition *data.Condition) bool {
	op := condition.Op
	// segment hasn't any operation
	if op == "" {
		op = condition.Property
	}
	if strings.Contains(op, ThanClause) {
		return thanCondition(user, condition)
	}
	switch op {
	case EqClause:
		return equalsCondition(user, condition)
	case NeqClause:
		return !equalsCondition(user, condition)
	case ContainsClause:
		return containsCondition(user, condition)
	case NotContainClause:
		return !containsCondition(user, condition)
	case IsOneOfClause:
		return oneOfCondition(user, condition)
	case NotOneOfClause:
		return !oneOfCondition(user, condition)
	case StartsWithClause:
		return startWithCondition(user, condition)
	case EndsWithClause:
		return endWithCondition(user, condition)
	case IsTrueClause:
		return trueCondition(user, condition)
	case IsFalseClause:
		return falseCondition(user, condition)
	case MatchRegexClause:
		return matchRegExCondition(user, condition)
	case NotMatchRegexClause:
		return !matchRegExCondition(user, condition)
	case IsInSegmentClause:
		return e.isInSegmentCondition(user, condition)
	case NotInSegmentClause:
		return !e.isInSegmentCondition(user, condition)
	}
	return false
}

func (e *evaluator) isInSegmentCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.GetKey()
	cv := condition.Value
	var segments []string
	if err := json.Unmarshal([]byte(cv), &segments); err != nil {
		return false
	}
	for _, sid := range segments {
		segment := e.getSegment(sid)
		if segment == nil {
			continue
		}
		switch segment.MatchUser(pv) {
		case data.SegmentExcludeUser:
		case data.SegmentIncludeUser:
			return true
		default:
			for _, rule := range segment.Rules {
				if e.ifUserMatchRule(user, rule.Conditions) {
					return true
				}
			}
		}
	}
	return false
}

func thanCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	var pvNumber, cvNumber float64
	var err error
	pvNumber, err = strconv.ParseFloat(pv, 64)
	if err != nil {
		return false
	}
	cvNumber, err = strconv.ParseFloat(cv, 64)
	if err != nil {
		return false
	}
	switch condition.Op {
	case GeClause:
		return pvNumber >= cvNumber
	case GtClause:
		return pvNumber > cvNumber
	case LeClause:
		return pvNumber <= cvNumber
	case LtClause:
		return pvNumber < cvNumber
	default:
		return false
	}
}

func equalsCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	return cv != "" && cv == pv
}

func containsCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	return pv != "" && cv != "" && strings.Contains(pv, cv)
}

func oneOfCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	if pv == "" {
		return false
	}
	cv := condition.Value
	var slice []string
	if err := json.Unmarshal([]byte(cv), &slice); err != nil {
		return false
	}
	for _, s := range slice {
		if s == pv {
			return true
		}
	}
	return false
}

func startWithCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	return pv != "" && cv != "" && strings.HasPrefix(pv, cv)
}

func endWithCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	return pv != "" && cv != "" && strings.HasSuffix(pv, cv)
}

func trueCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	return pv != "" && strings.EqualFold(pv, "true")
}

func falseCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	return pv != "" && strings.EqualFold(pv, "false")
}

func matchRegExCondition(user *FBUser, condition *data.Condition) bool {
	pv := user.Get(condition.Property)
	cv := condition.Value
	if pv == "" || cv == "" {
		return false
	}
	ret, _ := regexp.MatchString(cv, pv)
	return ret
}

func getRolloutVariationValue(flag *data.FeatureFlag,
	rollouts []data.RolloutVariation,
	user *FBUser,
	reason string,
	ruleIncludedInExperiment bool,
	dispatchKey string,
) (*evalResult, bool) {
	key := dispatchKey
	if key == "" {
		key = "keyid"
	}
	keyValue := user.Get(key)
	dispatchKeyValue := strings.Join([]string{flag.Key, keyValue}, "")
	var r *data.RolloutVariation
	for _, rollout := range rollouts {
		if util.IfKeyBelongsPercentage(dispatchKeyValue, rollout.Rollout) {
			r = &rollout
			break
		}
	}
	if r != nil {
		return &evalResult{
			id:               r.Id,
			reason:           reason,
			keyName:          flag.Key,
			name:             flag.Name,
			sendToExperiment: isSendToExperiment(dispatchKeyValue, r, flag.ExptIncludeAllTargets, ruleIncludedInExperiment),
			fv:               flag.GetFlagValue(r.Id),
			success:          true,
			flagType:         flag.VariationType,
		}, true
	}
	return nil, false
}

func isSendToExperiment(dispatchKey string, rollout *data.RolloutVariation, exptIncludeAllRules bool, ruleIncludedInExperiment bool) bool {
	if exptIncludeAllRules {
		return true
	}

	if ruleIncludedInExperiment {
		sendToExperimentPercentage := rollout.ExptRollout
		splittingPercentage := rollout.SplittingPercentage()
		if sendToExperimentPercentage == 0 || splittingPercentage == 0 {
			return false
		}
		upperBound := sendToExperimentPercentage / splittingPercentage
		if upperBound > 1 {
			upperBound = 1
		}
		exptDispatchKeyValue := strings.Join([]string{ExptKeyPrefix, dispatchKey}, "")
		return util.IfKeyBelongsPercentage(exptDispatchKeyValue, []float64{0, upperBound})
	}

	return false
}
