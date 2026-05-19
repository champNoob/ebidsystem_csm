package apperror

import "errors"

// BusinessError 业务错误结构体：
type BusinessError struct {
	Code    string
	Message string
	Cause   error
}

// 实现 Error 接口：
func (e *BusinessError) Error() string {
	return e.Message
}

// Unwrap 实现 Unwrap 接口：
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

func (e *BusinessError) WithCause(cause error) *BusinessError {
	return &BusinessError{
		Code:    e.Code,
		Message: e.Message,
		Cause:   cause,
	}
}

func AsBusinessError(err error) (*BusinessError, bool) {
	var be *BusinessError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}

func CodeOf(err error) string {
	if be, ok := AsBusinessError(err); ok {
		return be.Code
	}
	return ErrInternal.Code
}

func MessageOf(err error) string {
	if be, ok := AsBusinessError(err); ok {
		return be.Message
	}
	return ErrInternal.Message
}
