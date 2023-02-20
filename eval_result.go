package featbit

import (
	"encoding/json"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"reflect"
	"strconv"
)

type allFlagStateImpl struct {
	success   bool
	reason    string
	states    map[string]map[evalResult]*insight.FlagEvent
	sendEvent func(Event)
}

func (a allFlagStateImpl) IsSuccess() bool {
	return a.success
}

func (a allFlagStateImpl) Reason() string {
	return a.reason
}

func (a allFlagStateImpl) GetStringVariation(featureFlagKey string, defaultValue string) (string, EvalDetail, error) {
	ed, err := a.get(featureFlagKey, FlagStringType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret := ed.Variation.(string)
	return ret, ed, nil
}

func (a allFlagStateImpl) GetBoolVariation(featureFlagKey string, defaultValue bool) (bool, EvalDetail, error) {
	ed, err := a.get(featureFlagKey, FlagBoolType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret := ed.Variation.(bool)
	return ret, ed, nil
}

func (a allFlagStateImpl) GetIntVariation(featureFlagKey string, defaultValue int) (int, EvalDetail, error) {
	ed, err := a.get(featureFlagKey, FlagNumericType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret := ed.Variation.(int)
	return ret, ed, nil
}

func (a allFlagStateImpl) GetDoubleVariation(featureFlagKey string, defaultValue float64) (float64, EvalDetail, error) {
	ed, err := a.get(featureFlagKey, FlagNumericType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	ret := ed.Variation.(float64)
	return ret, ed, nil
}

func (a allFlagStateImpl) GetJsonVariation(featureFlagKey string, defaultValue interface{}) (interface{}, EvalDetail, error) {
	ed, err := a.get(featureFlagKey, FlagJsonType, defaultValue)
	if err != nil {
		return defaultValue, ed, err
	}
	return ed.Variation, ed, nil
}

func (a allFlagStateImpl) get(featureFlagKey string, requiredType string, defaultValue interface{}) (EvalDetail, error) {
	res, ok := a.states[featureFlagKey]
	if !ok {
		ed := EvalDetail{Variation: defaultValue, Reason: ReasonFlagNotFound, KeyName: featureFlagKey, Name: FlagNameUnknown}
		return ed, flagNotFound
	}
	for er, event := range res {
		if er.checkType(requiredType) {
			ed, err := er.castVariationByFlagType(requiredType, defaultValue)
			if err == nil && a.sendEvent != nil && event != nil {
				event.UpdateTimestamp()
				a.sendEvent(event)
			}
			return ed, err
		}
		return EvalDetail{Variation: defaultValue, Reason: er.reason, KeyName: er.keyName, Name: er.name}, evalWrongType
	}
	// impossible to reach here
	return EvalDetail{}, nil
}

type evalResult struct {
	id               string
	fv               string
	sendToExperiment bool
	success          bool
	flagType         string
	reason           string
	keyName          string
	name             string
}

func errorResult(reason string, keyName string, name string) *evalResult {
	return &evalResult{reason: reason, keyName: keyName, name: name}
}

func (er *evalResult) toEventFlag() insight.EventFlag {
	return insight.NewEventFlag(er.keyName, er.sendToExperiment, er.id, er.fv, er.reason)
}

func (er *evalResult) checkType(requiredType string) bool {
	switch er.flagType {
	case FlagBoolType:
		return requiredType == FlagBoolType || requiredType == FlagStringType
	case FlagNumericType:
		if requiredType == FlagBoolType {
			_, err := strconv.ParseFloat(er.fv, 64)
			return err == nil
		}
		return true
	case FlagJsonType, FlagStringType:
		if requiredType == FlagBoolType {
			_, err := strconv.ParseBool(er.fv)
			return err == nil
		}
		if requiredType == FlagNumericType {
			_, err := strconv.ParseFloat(er.fv, 64)
			return err == nil
		}
		return true
	}
	return false
}

func (er *evalResult) castVariationByFlagType(requiredType string, defaultValue interface{}) (EvalDetail, error) {
	switch requiredType {
	case FlagBoolType:
		b, _ := strconv.ParseBool(er.fv)
		return EvalDetail{Variation: b, Reason: er.reason, KeyName: er.keyName, Name: er.name}, nil
	case FlagNumericType:
		f, _ := strconv.ParseFloat(er.fv, 64)
		if reflect.TypeOf(defaultValue).Kind() == reflect.Int {
			return EvalDetail{Variation: int(f), Reason: er.reason, KeyName: er.keyName, Name: er.name}, nil
		}
		return EvalDetail{Variation: f, Reason: er.reason, KeyName: er.keyName, Name: er.name}, nil

	case FlagJsonType:
		t := reflect.TypeOf(defaultValue)
		inf := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(er.fv), inf); err != nil {
			log.LogError("FB GO SDK: unexpected error in parsing json, use default value")
			return EvalDetail{Variation: defaultValue, Reason: er.reason, KeyName: er.keyName, Name: er.name}, err
		}
		inf = reflect.ValueOf(inf).Elem().Interface()
		return EvalDetail{Variation: inf, Reason: er.reason, KeyName: er.keyName, Name: er.name}, nil
	default:
		return EvalDetail{Variation: er.fv, Reason: er.reason, KeyName: er.keyName, Name: er.name}, nil
	}

}
