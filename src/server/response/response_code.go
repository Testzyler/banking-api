package response

type ResponseCode uint64

const baseResponseCode uint64 = 10000

func newResponseCode(code uint64) ResponseCode {
	return ResponseCode(baseResponseCode + code)
}

var (
	// Success codes
	SystemOperationSuccess = newResponseCode(0)
	SuccessCodeOK          = newResponseCode(200)

	// 4xx Client Error codes
	ErrCodeNotFound         = newResponseCode(404)
	ErrCodeBadRequest       = newResponseCode(400)
	ErrCodeUnauthorized     = newResponseCode(401)
	ErrCodeForbidden        = newResponseCode(403)
	ErrCodeValidationFailed = newResponseCode(422)

	// >5xx Server Error codes
	ErrCodeInternalServer     = newResponseCode(500)
	ErrCodeServiceUnavailable = newResponseCode(503)
	ErrCodeDatabaseError      = newResponseCode(600)
	ErrCodeUnknownError       = newResponseCode(700)
)

var CustomResponseMessages = map[ResponseCode]string{
	// Success codes
	SystemOperationSuccess: "Request completed successfully",

	// Client Error codes
	ErrCodeNotFound:         "Not Found",
	ErrCodeBadRequest:       "Bad Request",
	ErrCodeUnauthorized:     "Unauthorized",
	ErrCodeForbidden:        "Forbidden",
	ErrCodeValidationFailed: "Validation Failed",

	// Server Error codes
	ErrCodeInternalServer:     "Internal Server Error",
	ErrCodeServiceUnavailable: "Service Unavailable",
	ErrCodeDatabaseError:      "Database Error",
	ErrCodeUnknownError:       "Unknown Error",
}
