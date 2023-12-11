package proxy_factory

import (
	"fmt"
	"reflect"

	perrors "github.com/pkg/errors"
)

// CallLocalMethod is used to handle invoke exception in user func.
func callLocalMethod(method reflect.Method, in []reflect.Value) ([]reflect.Value, error) {
	var (
		returnValues []reflect.Value
		retErr       error
	)

	func() {
		defer func() {
			if e := recover(); e != nil {
				if err, ok := e.(error); ok {
					retErr = err
				} else if err, ok := e.(string); ok {
					retErr = perrors.New(err)
				} else {
					retErr = fmt.Errorf("invoke function error, unknow exception: %+v", e)
				}
			}
		}()
		returnValues = method.Func.Call(in)
	}()

	if retErr != nil {
		return nil, retErr
	}

	return returnValues, retErr
}
