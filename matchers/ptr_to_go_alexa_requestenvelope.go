// Code generated by pegomock. DO NOT EDIT.
package matchers

import (
	"github.com/petergtz/pegomock"
	"reflect"

	go_alexa "github.com/petergtz/go-alexa"
)

func AnyPtrToGoAlexaRequestEnvelope() *go_alexa.RequestEnvelope {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf((*(*go_alexa.RequestEnvelope))(nil)).Elem()))
	var nullValue *go_alexa.RequestEnvelope
	return nullValue
}

func EqPtrToGoAlexaRequestEnvelope(value *go_alexa.RequestEnvelope) *go_alexa.RequestEnvelope {
	pegomock.RegisterMatcher(&pegomock.EqMatcher{Value: value})
	var nullValue *go_alexa.RequestEnvelope
	return nullValue
}

func NotEqPtrToGoAlexaRequestEnvelope(value *go_alexa.RequestEnvelope) *go_alexa.RequestEnvelope {
	pegomock.RegisterMatcher(&pegomock.NotEqMatcher{Value: value})
	var nullValue *go_alexa.RequestEnvelope
	return nullValue
}

func PtrToGoAlexaRequestEnvelopeThat(matcher pegomock.ArgumentMatcher) *go_alexa.RequestEnvelope {
	pegomock.RegisterMatcher(matcher)
	var nullValue *go_alexa.RequestEnvelope
	return nullValue
}
