package apperrors

import (
	"context"
	"errors"
	"reflect"
)

// IsClientCancellation 判断错误树中所有非 nil 原因是否都为请求取消。
// 包装或组合中含有超时、真实异常时仍须保留日志。
func IsClientCancellation(err error) bool {
	if err == nil {
		return false
	}
	// 类型化 nil 仍是非 nil 接口，调用其方法可能 panic；保留日志并沿用原错误编码。
	value := reflect.ValueOf(err)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return false
		}
	}

	switch wrapped := err.(type) {
	case interface{ Unwrap() error }:
		if cause := wrapped.Unwrap(); cause != nil {
			return IsClientCancellation(cause)
		}
	case interface{ Unwrap() []error }:
		hasCause := false
		for _, cause := range wrapped.Unwrap() {
			if cause == nil {
				continue
			}
			hasCause = true
			if !IsClientCancellation(cause) {
				return false
			}
		}
		if hasCause {
			return true
		}
	}

	return errors.Is(err, context.Canceled)
}
