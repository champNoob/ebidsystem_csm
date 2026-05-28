package apperror

import "errors"

// BusinessError 业务错误结构体：
type BusinessError struct {
	Code    string
	Message string
	Cause   error
}

// 返回错误消息：
func (e *BusinessError) Error() string {
	return e.Message
}

// 返回底层的 Cause 错误，用于错误链：
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// 返回一个新的 BusinessError 实例：
func (e *BusinessError) WithCause(cause error) *BusinessError {
	return &BusinessError{
		Code:    e.Code,
		Message: e.Message,
		Cause:   cause,
	}
}

// Wrap 用指定的业务错误包装底层原因：
/* 对外仍然暴露 base 的 Code 和 Message，
底层 cause 通过 errors.Unwrap / errors.As 保留，供日志和调试使用 */
func Wrap(base *BusinessError, cause error) error {
	if cause == nil {
		return base
	}
	if base == nil {
		return cause
	}
	return base.WithCause(cause)
}

// 类型断言：
func AsBusinessError(err error) (*BusinessError, bool) {
	var be *BusinessError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

// 提取错误码：
func CodeOf(err error) string {
	if be, ok := AsBusinessError(err); ok {
		return be.Code
	}
	return ErrInternal.Code
}

// 提取错误消息：
func MessageOf(err error) string {
	if be, ok := AsBusinessError(err); ok {
		return be.Message
	}
	return ErrInternal.Message
}
