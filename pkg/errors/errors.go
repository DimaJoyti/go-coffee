package errors

import (
	"fmt"
	"runtime"
)

// AppError представляє помилку додатку
type AppError struct {
	// Err оригінальна помилка
	Err error
	// Message повідомлення про помилку
	Message string
	// Code код помилки
	Code string
	// StatusCode HTTP статус код
	StatusCode int
	// Stack стек виклику
	Stack string
	// Context контекст помилки
	Context map[string]interface{}
}

// Error повертає повідомлення про помилку
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap повертає оригінальну помилку
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext додає контекст до помилки
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New створює нову помилку
func New(message string) *AppError {
	return &AppError{
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap обгортає помилку
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// Якщо помилка вже є AppError, додаємо повідомлення
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Err:        appErr.Err,
			Message:    fmt.Sprintf("%s: %s", message, appErr.Message),
			Code:       appErr.Code,
			StatusCode: appErr.StatusCode,
			Stack:      appErr.Stack,
			Context:    appErr.Context,
		}
	}

	return &AppError{
		Err:     err,
		Message: message,
		Stack:   getStack(),
	}
}

// WithCode додає код помилки
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithStatusCode додає HTTP статус код
func (e *AppError) WithStatusCode(statusCode int) *AppError {
	e.StatusCode = statusCode
	return e
}

// getStack повертає стек виклику
func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack string
	for {
		frame, more := frames.Next()
		stack += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return stack
}
