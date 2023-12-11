package invocation

import (
	"context"
	"reflect"
	"sync"

	"google.golang.org/grpc/metadata"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/protocol"
)

var _ protocol.Invocation = (*RPCInvocation)(nil)

// nolint
type RPCInvocation struct {
	methodName string
	// Parameter Type Names. It is used to specify the parameterType
	parameterTypeNames []string
	parameterTypes     []reflect.Type

	parameterValues []reflect.Value
	arguments       []interface{}
	reply           interface{}
	callBack        interface{}
	attachments     map[string]interface{}

	attributes map[string]interface{}
	invoker    protocol.Invoker
	lock       sync.RWMutex
}

// NewRPCInvocation creates a RPC invocation.
func NewRPCInvocation(methodName string, arguments []interface{}, attachments map[string]interface{}) *RPCInvocation {
	return &RPCInvocation{
		methodName:  methodName,
		arguments:   arguments,
		attachments: attachments,
		attributes:  make(map[string]interface{}, 8),
	}
}

// NewRPCInvocationWithOptions creates a RPC invocation with @opts.
func NewRPCInvocationWithOptions(opts ...option) *RPCInvocation {
	invo := &RPCInvocation{}
	for _, opt := range opts {
		opt(invo)
	}
	if invo.attributes == nil {
		invo.attributes = make(map[string]interface{})
	}
	return invo
}

// MethodName gets RPC invocation method name.
func (r *RPCInvocation) MethodName() string {
	return r.methodName
}

// ActualMethodName gets actual invocation method name. It returns the method name been called if it's a generic call
func (r *RPCInvocation) ActualMethodName() string {
	return r.MethodName()
}

// ParameterTypes gets RPC invocation parameter types.
func (r *RPCInvocation) ParameterTypes() []reflect.Type {
	return r.parameterTypes
}

// ParameterTypeNames gets RPC invocation parameter types of string expression.
func (r *RPCInvocation) ParameterTypeNames() []string {
	return r.parameterTypeNames
}

// ParameterValues gets RPC invocation parameter values.
func (r *RPCInvocation) ParameterValues() []reflect.Value {
	return r.parameterValues
}

// Arguments gets RPC arguments.
func (r *RPCInvocation) Arguments() []interface{} {
	return r.arguments
}

// Reply gets response of RPC request.
func (r *RPCInvocation) Reply() interface{} {
	return r.reply
}

// SetReply sets response of RPC request.
func (r *RPCInvocation) SetReply(reply interface{}) {
	r.reply = reply
}

// Attachments gets all attachments of RPC.
func (r *RPCInvocation) Attachments() map[string]interface{} {
	return r.attachments
}

// GetAttachmentInterface returns the corresponding value from service's attachment with the given key.
func (r *RPCInvocation) GetAttachmentInterface(key string) interface{} {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.attachments == nil {
		return nil
	}
	return r.attachments[key]
}

// Attributes gets all attributes of RPC.
func (r *RPCInvocation) Attributes() map[string]interface{} {
	return r.attributes
}

// Invoker gets the invoker in current context.
func (r *RPCInvocation) Invoker() protocol.Invoker {
	return r.invoker
}

// nolint
func (r *RPCInvocation) SetInvoker(invoker protocol.Invoker) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.invoker = invoker
}

// CallBack sets RPC callback method.
func (r *RPCInvocation) CallBack() interface{} {
	return r.callBack
}

// SetCallBack sets RPC callback method.
func (r *RPCInvocation) SetCallBack(c interface{}) {
	r.callBack = c
}

func (r *RPCInvocation) ServiceKey() string {
	return common.ServiceKey(
		r.GetAttachmentWithDefaultValue(constant.InterfaceKey, ""),
		r.GetAttachmentWithDefaultValue(constant.GroupKey, ""),
		r.GetAttachmentWithDefaultValue(constant.VersionKey, ""))
}

func (r *RPCInvocation) SetAttachment(key string, value interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.attachments == nil {
		r.attachments = make(map[string]interface{})
	}
	r.attachments[key] = value
}

func (r *RPCInvocation) GetAttachment(key string) (string, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.attachments == nil {
		return "", false
	}
	if value, existed := r.attachments[key]; existed {
		if str, strOK := value.(string); strOK {
			return str, true
		} else if strArr, strArrOK := value.([]string); strArrOK && len(strArr) > 0 {
			// For triple protocol, the attachment value is wrapped by an array.
			return strArr[0], true
		}
	}
	return "", false
}

func (r *RPCInvocation) GetAttachmentWithDefaultValue(key string, defaultValue string) string {
	if value, ok := r.GetAttachment(key); ok {
		return value
	}
	return defaultValue
}

func (r *RPCInvocation) SetAttribute(key string, value interface{}) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.attributes == nil {
		r.attributes = make(map[string]interface{})
	}
	r.attributes[key] = value
}

func (r *RPCInvocation) GetAttribute(key string) (interface{}, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.attributes == nil {
		return nil, false
	}
	value, ok := r.attributes[key]
	return value, ok
}

func (r *RPCInvocation) GetAttributeWithDefaultValue(key string, defaultValue interface{}) interface{} {
	r.lock.RLock()
	defer r.lock.RUnlock()
	if r.attributes == nil {
		return defaultValue
	}
	if value, ok := r.attributes[key]; ok {
		return value
	}
	return defaultValue
}

func (r *RPCInvocation) GetAttachmentAsContext() context.Context {
	gRPCMD := make(metadata.MD, 0)
	ctx := context.Background()
	for k, v := range r.Attachments() {
		if str, ok := v.(string); ok {
			gRPCMD.Set(k, str)
			continue
		}
		if str, ok := v.([]string); ok {
			gRPCMD.Set(k, str...)
			continue
		}
	}
	return metadata.NewOutgoingContext(ctx, gRPCMD)
}

// /////////////////////////
// option
// /////////////////////////

type option func(invo *RPCInvocation)

// WithMethodName creates option with @methodName.
func WithMethodName(methodName string) option {
	return func(invo *RPCInvocation) {
		invo.methodName = methodName
	}
}

// WithParameterTypes creates option with @parameterTypes.
func WithParameterTypes(parameterTypes []reflect.Type) option {
	return func(invo *RPCInvocation) {
		invo.parameterTypes = parameterTypes
	}
}

// WithParameterTypeNames creates option with @parameterTypeNames.
func WithParameterTypeNames(parameterTypeNames []string) option {
	return func(invo *RPCInvocation) {
		if len(parameterTypeNames) == 0 {
			return
		}
		parameterTypeNamesTmp := make([]string, len(parameterTypeNames))
		copy(parameterTypeNamesTmp, parameterTypeNames)
		invo.parameterTypeNames = parameterTypeNamesTmp
	}
}

// WithParameterValues creates option with @parameterValues
func WithParameterValues(parameterValues []reflect.Value) option {
	return func(invo *RPCInvocation) {
		invo.parameterValues = parameterValues
	}
}

// WithArguments creates option with @arguments function.
func WithArguments(arguments []interface{}) option {
	return func(invo *RPCInvocation) {
		invo.arguments = arguments
	}
}

// WithReply creates option with @reply function.
func WithReply(reply interface{}) option {
	return func(invo *RPCInvocation) {
		invo.reply = reply
	}
}

// WithCallBack creates option with @callback function.
func WithCallBack(callBack interface{}) option {
	return func(invo *RPCInvocation) {
		invo.callBack = callBack
	}
}

// WithAttachments creates option with @attachments.
func WithAttachments(attachments map[string]interface{}) option {
	return func(invo *RPCInvocation) {
		invo.attachments = attachments
	}
}

// WithAttachment put a key-value pair into attachments.
func WithAttachment(k string, v interface{}) option {
	return func(invo *RPCInvocation) {
		if invo.attachments == nil {
			invo.attachments = make(map[string]interface{})
		}
		invo.attachments[k] = v
	}
}

// WithInvoker creates option with @invoker.
func WithInvoker(invoker protocol.Invoker) option {
	return func(invo *RPCInvocation) {
		invo.invoker = invoker
	}
}
