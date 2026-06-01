package errors

import "fmt"

type ErrorCode string

const (
	ErrInvalidRequest ErrorCode = "INVALID_REQUEST"
	ErrNotFound      ErrorCode = "NOT_FOUND"
	ErrInternal      ErrorCode = "INTERNAL_ERROR"
	ErrKafkaError    ErrorCode = "KAFKA_ERROR"
	ErrRedisError    ErrorCode = "REDIS_ERROR"
	ErrSendGridError ErrorCode = "SENDGRID_ERROR"
	ErrTwilioError   ErrorCode = "TWILIO_ERROR"
	ErrFCMError      ErrorCode = "FCM_ERROR"
	ErrTimeout       ErrorCode = "TIMEOUT"
	ErrUnauthorized  ErrorCode = "UNAUTHORIZED"
)

type ServiceError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

func NewErrorWithErr(code ErrorCode, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *ServiceError) StatusCode() int {
	switch e.Code {
	case ErrInvalidRequest:
		return 400
	case ErrNotFound:
		return 404
	case ErrUnauthorized:
		return 401
	case ErrTimeout:
		return 408
	default:
		return 500
	}
}
