package exceptions

type ErrorCode int

const baseErrorCode int = 10000

func newErrorCode(code int) ErrorCode {
	return ErrorCode(baseErrorCode + code)
}

var (
	// 4xx client errors
	ErrCodeNotFound       = newErrorCode(404)
	ErrCodeInvalidRequest = newErrorCode(400)
	ErrCodeUnauthorized   = newErrorCode(401)
	ErrCodeForbidden      = newErrorCode(403)
	ErrValidationFailed   = newErrorCode(422)

	// 5xx server errors
	ErrCodeInternalServer     = newErrorCode(500)
	ErrCodeServiceUnavailable = newErrorCode(503)
	ErrCodeDatabaseError      = newErrorCode(600)
	ErrUnknownError           = newErrorCode(700)
)
