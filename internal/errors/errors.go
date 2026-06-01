package errors

import (
	"fmt"
)

type MemoError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (e *MemoError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	if len(e.Context) > 0 {
		msg += fmt.Sprintf(" %v", e.Context)
	}
	return msg
}

func (e *MemoError) Unwrap() error {
	return e.Cause
}

func (e *MemoError) Is(target error) bool {
	t, ok := target.(*MemoError)
	return ok && t.Code == e.Code
}

func (e *MemoError) With(key string, value interface{}) *MemoError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func InvalidInput(message string, fields ...string) *MemoError {
	err := &MemoError{
		Code:    ErrCodeInvalidInput,
		Message: message,
	}
	if len(fields) > 0 {
		err.With("fields", fields)
	}
	return err
}

func NotFound(resource string, id interface{}) *MemoError {
	return &MemoError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Context: map[string]interface{}{"id": id},
	}
}

func DatabaseError(err error, operation string) *MemoError {
	return &MemoError{
		Code:    ErrCodeDatabase,
		Message: fmt.Sprintf("database error during %s", operation),
		Cause:   err,
		Context: map[string]interface{}{"operation": operation},
	}
}

func EmbeddingError(err error, provider string) *MemoError {
	return &MemoError{
		Code:    ErrCodeEmbedding,
		Message: "failed to generate embedding",
		Cause:   err,
		Context: map[string]interface{}{"provider": provider},
	}
}

func ProjectError(err error, cwd string) *MemoError {
	return &MemoError{
		Code:    ErrCodeProject,
		Message: "failed to resolve project",
		Cause:   err,
		Context: map[string]interface{}{"cwd": cwd},
	}
}

func ConfigError(err error) *MemoError {
	return &MemoError{
		Code:    ErrCodeConfig,
		Message: "configuration error",
		Cause:   err,
	}
}

func InternalError(message string) *MemoError {
	return &MemoError{
		Code:    ErrCodeInternal,
		Message: message,
	}
}

func UnhealthyError(component string, err error) *MemoError {
	return &MemoError{
		Code:    ErrCodeUnhealthy,
		Message: fmt.Sprintf("%s is unhealthy", component),
		Cause:   err,
		Context: map[string]interface{}{"component": component},
	}
}

func ConflictError(resource string, message string) *MemoError {
	return &MemoError{
		Code:    ErrCodeConflict,
		Message: fmt.Sprintf("%s conflict: %s", resource, message),
	}
}

func TimeoutError(operation string) *MemoError {
	return &MemoError{
		Code:    ErrCodeTimeout,
		Message: fmt.Sprintf("%s timed out", operation),
		Context: map[string]interface{}{"operation": operation},
	}
}
