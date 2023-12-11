package protocol

import (
	"context"
	"reflect"
)

// Invocation is a interface which is invocation for each remote method.
type Invocation interface {
	// MethodName gets invocation method name.
	MethodName() string

	// ActualMethodName gets actual invocation method name. It returns the method name been called if it's a generic call
	ActualMethodName() string

	// ParameterTypeNames gets invocation parameter type names.
	ParameterTypeNames() []string

	// ParameterTypes gets invocation parameter types.
	ParameterTypes() []reflect.Type

	// ParameterValues gets invocation parameter values.
	ParameterValues() []reflect.Value

	// Arguments gets arguments.
	Arguments() []interface{}

	// Reply gets response of request
	Reply() interface{}

	// Invoker gets the invoker in current context.
	Invoker() Invoker

	// Attachments gets all attachments
	Attachments() map[string]interface{}
	SetAttachment(key string, value interface{})
	GetAttachment(key string) (string, bool)
	GetAttachmentInterface(string) interface{}
	GetAttachmentWithDefaultValue(key string, defaultValue string) string
	GetAttachmentAsContext() context.Context

	Attributes() map[string]interface{}
	SetAttribute(key string, value interface{})
	GetAttribute(key string) (interface{}, bool)
	GetAttributeWithDefaultValue(key string, defaultValue interface{}) interface{}
}
