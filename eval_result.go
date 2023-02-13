package featbit

import (
	"encoding/json"
	. "github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"reflect"
	"strconv"
)

type detail struct {
	Reason  string `json:"reason"`
	KeyName string `json:"keyName"`
	Name    string `json:"name"`
}

type evalDetailImpl struct {
	Variation interface{} `json:"variation"`
	detail
}

func (ed evalDetailImpl) GetVariation() interface{} {
	return ed.Variation
}

func (ed evalDetailImpl) GetReason() string {
	return ed.Reason
}

func (ed evalDetailImpl) GetKeyName() string {
	return ed.KeyName
}

func (ed evalDetailImpl) GetName() string {
	return ed.Name
}

type evalResult struct {
	id               string
	fv               string
	sendToExperiment bool
	success          bool
	flagType         string
	detail
}

func errorResult(reason string, keyName string, name string) evalResult {
	return evalResult{
		detail: detail{
			Reason:  reason,
			KeyName: keyName,
			Name:    name,
		},
	}
}

func (er *evalResult) toEventFlag() insight.EventFlag {
	return insight.NewEventFlag(er.KeyName, er.sendToExperiment, er.id, er.fv, er.Reason)
}

func (er *evalResult) checkType(requiredType string) bool {
	switch er.flagType {
	case FlagBoolType:
		return requiredType == FlagBoolType
	case FlagNumericType:
		_, err := strconv.ParseFloat(er.fv, 64)
		return err == nil
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
		return evalDetailImpl{Variation: b, detail: er.detail}, nil
	case FlagNumericType:
		if reflect.TypeOf(defaultValue).Kind() == reflect.Int {
			i, _ := strconv.Atoi(er.fv)
			return evalDetailImpl{Variation: i, detail: er.detail}, nil
		} else {
			f, _ := strconv.ParseFloat(er.fv, 64)
			return evalDetailImpl{Variation: f, detail: er.detail}, nil
		}
	case FlagJsonType:
		t := reflect.TypeOf(defaultValue)
		inf := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(er.fv), inf); err != nil {
			log.LogError("FB GO SDK: unexpected error in parsing json, use default value")
			return evalDetailImpl{Variation: defaultValue, detail: er.detail}, err
		}
		inf = reflect.ValueOf(inf).Elem().Interface()
		return evalDetailImpl{Variation: inf, detail: er.detail}, nil
	default:
		return evalDetailImpl{Variation: er.fv, detail: er.detail}, nil
	}

}
