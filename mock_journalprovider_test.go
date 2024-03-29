// Code generated by pegomock. DO NOT EDIT.
// Source: github.com/petergtz/alexa-journal (interfaces: JournalProvider)

package journalskill_test

import (
	journal "github.com/petergtz/alexa-journal/journal"
	pegomock "github.com/petergtz/pegomock"
	"reflect"
	"time"
)

type MockJournalProvider struct {
	fail func(message string, callerSkip ...int)
}

func NewMockJournalProvider(options ...pegomock.Option) *MockJournalProvider {
	mock := &MockJournalProvider{}
	for _, option := range options {
		option.Apply(mock)
	}
	return mock
}

func (mock *MockJournalProvider) SetFailHandler(fh pegomock.FailHandler) { mock.fail = fh }
func (mock *MockJournalProvider) FailHandler() pegomock.FailHandler      { return mock.fail }

func (mock *MockJournalProvider) Get(accessToken string, spreadsheetName string) (journal.Journal, error) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockJournalProvider().")
	}
	params := []pegomock.Param{accessToken, spreadsheetName}
	result := pegomock.GetGenericMockFrom(mock).Invoke("Get", params, []reflect.Type{reflect.TypeOf((*journal.Journal)(nil)).Elem(), reflect.TypeOf((*error)(nil)).Elem()})
	var ret0 journal.Journal
	var ret1 error
	if len(result) != 0 {
		if result[0] != nil {
			ret0 = result[0].(journal.Journal)
		}
		if result[1] != nil {
			ret1 = result[1].(error)
		}
	}
	return ret0, ret1
}

func (mock *MockJournalProvider) VerifyWasCalledOnce() *VerifierMockJournalProvider {
	return &VerifierMockJournalProvider{
		mock:                   mock,
		invocationCountMatcher: pegomock.Times(1),
	}
}

func (mock *MockJournalProvider) VerifyWasCalled(invocationCountMatcher pegomock.InvocationCountMatcher) *VerifierMockJournalProvider {
	return &VerifierMockJournalProvider{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
	}
}

func (mock *MockJournalProvider) VerifyWasCalledInOrder(invocationCountMatcher pegomock.InvocationCountMatcher, inOrderContext *pegomock.InOrderContext) *VerifierMockJournalProvider {
	return &VerifierMockJournalProvider{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		inOrderContext:         inOrderContext,
	}
}

func (mock *MockJournalProvider) VerifyWasCalledEventually(invocationCountMatcher pegomock.InvocationCountMatcher, timeout time.Duration) *VerifierMockJournalProvider {
	return &VerifierMockJournalProvider{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		timeout:                timeout,
	}
}

type VerifierMockJournalProvider struct {
	mock                   *MockJournalProvider
	invocationCountMatcher pegomock.InvocationCountMatcher
	inOrderContext         *pegomock.InOrderContext
	timeout                time.Duration
}

func (verifier *VerifierMockJournalProvider) Get(accessToken string, spreadsheetName string) *MockJournalProvider_Get_OngoingVerification {
	params := []pegomock.Param{accessToken, spreadsheetName}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Get", params, verifier.timeout)
	return &MockJournalProvider_Get_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type MockJournalProvider_Get_OngoingVerification struct {
	mock              *MockJournalProvider
	methodInvocations []pegomock.MethodInvocation
}

func (c *MockJournalProvider_Get_OngoingVerification) GetCapturedArguments() (string, string) {
	accessToken, spreadsheetName := c.GetAllCapturedArguments()
	return accessToken[len(accessToken)-1], spreadsheetName[len(spreadsheetName)-1]
}

func (c *MockJournalProvider_Get_OngoingVerification) GetAllCapturedArguments() (_param0 []string, _param1 []string) {
	params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(params) > 0 {
		_param0 = make([]string, len(c.methodInvocations))
		for u, param := range params[0] {
			_param0[u] = param.(string)
		}
		_param1 = make([]string, len(c.methodInvocations))
		for u, param := range params[1] {
			_param1[u] = param.(string)
		}
	}
	return
}
