package errors

// ErrorCode represents a specific error code
type ErrorCode string

// Validation error codes (1000-1999)
const (
	CodeValidationFailed        ErrorCode = "E1001"
	CodeInvalidInput           ErrorCode = "E1002"
	CodeMissingRequiredField   ErrorCode = "E1003"
	CodeInvalidFormat          ErrorCode = "E1004"
	CodeValueOutOfRange        ErrorCode = "E1005"
	CodeInvalidLength          ErrorCode = "E1006"
	CodeInvalidCharacters      ErrorCode = "E1007"
	CodeDuplicateValue         ErrorCode = "E1008"
	CodeInvalidEnum            ErrorCode = "E1009"
	CodeInvalidUUID            ErrorCode = "E1010"
)

// Authentication error codes (2000-2999)
const (
	CodeAuthenticationFailed   ErrorCode = "E2001"
	CodeInvalidCredentials     ErrorCode = "E2002"
	CodeTokenExpired           ErrorCode = "E2003"
	CodeTokenInvalid           ErrorCode = "E2004"
	CodeTokenMissing           ErrorCode = "E2005"
	CodeAccountLocked          ErrorCode = "E2006"
	CodeAccountDisabled        ErrorCode = "E2007"
	CodePasswordExpired        ErrorCode = "E2008"
	CodeMFARequired            ErrorCode = "E2009"
	CodeMFAFailed              ErrorCode = "E2010"
)

// Authorization error codes (3000-3999)
const (
	CodeAuthorizationFailed    ErrorCode = "E3001"
	CodeInsufficientPermissions ErrorCode = "E3002"
	CodeAccessDenied           ErrorCode = "E3003"
	CodeResourceForbidden      ErrorCode = "E3004"
	CodeOperationNotAllowed    ErrorCode = "E3005"
	CodeQuotaExceeded          ErrorCode = "E3006"
	CodeRateLimitExceeded      ErrorCode = "E3007"
	CodeIPBlocked              ErrorCode = "E3008"
	CodeRegionRestricted       ErrorCode = "E3009"
	CodeMaintenanceMode        ErrorCode = "E3010"
)

// Resource error codes (4000-4999)
const (
	CodeResourceNotFound       ErrorCode = "E4001"
	CodeResourceAlreadyExists  ErrorCode = "E4002"
	CodeResourceConflict       ErrorCode = "E4003"
	CodeResourceLocked         ErrorCode = "E4004"
	CodeResourceCorrupted      ErrorCode = "E4005"
	CodeResourceTooLarge       ErrorCode = "E4006"
	CodeResourceExpired        ErrorCode = "E4007"
	CodeResourceUnavailable    ErrorCode = "E4008"
	CodeResourceDeleted        ErrorCode = "E4009"
	CodeResourceVersionMismatch ErrorCode = "E4010"
)

// Network error codes (5000-5999)
const (
	CodeNetworkError           ErrorCode = "E5001"
	CodeConnectionFailed       ErrorCode = "E5002"
	CodeConnectionTimeout      ErrorCode = "E5003"
	CodeConnectionReset        ErrorCode = "E5004"
	CodeDNSResolutionFailed    ErrorCode = "E5005"
	CodeSSLHandshakeFailed     ErrorCode = "E5006"
	CodeProxyError             ErrorCode = "E5007"
	CodeNetworkUnreachable     ErrorCode = "E5008"
	CodeHostUnreachable        ErrorCode = "E5009"
	CodePortUnreachable        ErrorCode = "E5010"
)

// Database error codes (6000-6999)
const (
	CodeDatabaseError          ErrorCode = "E6001"
	CodeDatabaseConnectionFailed ErrorCode = "E6002"
	CodeDatabaseTimeout        ErrorCode = "E6003"
	CodeDatabaseLockTimeout    ErrorCode = "E6004"
	CodeDatabaseDeadlock       ErrorCode = "E6005"
	CodeDatabaseConstraintViolation ErrorCode = "E6006"
	CodeDatabaseQueryError     ErrorCode = "E6007"
	CodeDatabaseTransactionFailed ErrorCode = "E6008"
	CodeDatabaseMigrationFailed ErrorCode = "E6009"
	CodeDatabaseCorruption     ErrorCode = "E6010"
)

// Messaging error codes (7000-7999)
const (
	CodeMessagingError         ErrorCode = "E7001"
	CodeMessagePublishFailed   ErrorCode = "E7002"
	CodeMessageConsumeFailed   ErrorCode = "E7003"
	CodeTopicNotFound          ErrorCode = "E7004"
	CodePartitionError         ErrorCode = "E7005"
	CodeOffsetError            ErrorCode = "E7006"
	CodeSerializationError     ErrorCode = "E7007"
	CodeDeserializationError   ErrorCode = "E7008"
	CodeMessageTooLarge        ErrorCode = "E7009"
	CodeBrokerUnavailable      ErrorCode = "E7010"
)

// AI service error codes (8000-8999)
const (
	CodeAIServiceError         ErrorCode = "E8001"
	CodeAIProviderUnavailable  ErrorCode = "E8002"
	CodeAIModelNotFound        ErrorCode = "E8003"
	CodeAITokenLimitExceeded   ErrorCode = "E8004"
	CodeAIRateLimitExceeded    ErrorCode = "E8005"
	CodeAIQuotaExceeded        ErrorCode = "E8006"
	CodeAIInvalidPrompt        ErrorCode = "E8007"
	CodeAIContentFiltered      ErrorCode = "E8008"
	CodeAIGenerationFailed     ErrorCode = "E8009"
	CodeAITimeout              ErrorCode = "E8010"
)

// External service error codes (9000-9999)
const (
	CodeExternalServiceError   ErrorCode = "E9001"
	CodeExternalServiceTimeout ErrorCode = "E9002"
	CodeExternalServiceUnavailable ErrorCode = "E9003"
	CodeExternalAPIError       ErrorCode = "E9004"
	CodeExternalAuthFailed     ErrorCode = "E9005"
	CodeExternalRateLimited    ErrorCode = "E9006"
	CodeExternalBadResponse    ErrorCode = "E9007"
	CodeExternalVersionMismatch ErrorCode = "E9008"
	CodeExternalConfigError    ErrorCode = "E9009"
	CodeExternalDependencyFailed ErrorCode = "E9010"
)

// Internal system error codes (10000-10999)
const (
	CodeInternalError          ErrorCode = "E10001"
	CodeConfigurationError     ErrorCode = "E10002"
	CodeInitializationFailed   ErrorCode = "E10003"
	CodeShutdownFailed         ErrorCode = "E10004"
	CodeMemoryError            ErrorCode = "E10005"
	CodeDiskSpaceError         ErrorCode = "E10006"
	CodeCPUOverload            ErrorCode = "E10007"
	CodeGoroutineLeak          ErrorCode = "E10008"
	CodePanicRecovered         ErrorCode = "E10009"
	CodeUnexpectedState        ErrorCode = "E10010"
)

// Timeout error codes (11000-11999)
const (
	CodeTimeout                ErrorCode = "E11001"
	CodeRequestTimeout         ErrorCode = "E11002"
	CodeOperationTimeout       ErrorCode = "E11003"
	CodeContextCancelled       ErrorCode = "E11004"
	CodeContextDeadlineExceeded ErrorCode = "E11005"
	CodeReadTimeout            ErrorCode = "E11006"
	CodeWriteTimeout           ErrorCode = "E11007"
	CodeIdleTimeout            ErrorCode = "E11008"
	CodeKeepAliveTimeout       ErrorCode = "E11009"
	CodeGracefulShutdownTimeout ErrorCode = "E11010"
)

// Circuit breaker error codes (12000-12999)
const (
	CodeCircuitBreakerOpen     ErrorCode = "E12001"
	CodeCircuitBreakerHalfOpen ErrorCode = "E12002"
	CodeCircuitBreakerFailed   ErrorCode = "E12003"
	CodeCircuitBreakerTimeout  ErrorCode = "E12004"
	CodeCircuitBreakerConfigError ErrorCode = "E12005"
	CodeTooManyFailures        ErrorCode = "E12006"
	CodeFailureThresholdExceeded ErrorCode = "E12007"
	CodeSuccessThresholdNotMet ErrorCode = "E12008"
	CodeCircuitBreakerReset    ErrorCode = "E12009"
	CodeCircuitBreakerTripped  ErrorCode = "E12010"
)

// Business logic error codes (13000-13999)
const (
	CodeBusinessRuleViolation  ErrorCode = "E13001"
	CodeWorkflowError          ErrorCode = "E13002"
	CodeStateTransitionError   ErrorCode = "E13003"
	CodeInvalidOperation       ErrorCode = "E13004"
	CodePreconditionFailed     ErrorCode = "E13005"
	CodePostconditionFailed    ErrorCode = "E13006"
	CodeInvariantViolation     ErrorCode = "E13007"
	CodeBusinessLogicError     ErrorCode = "E13008"
	CodeDomainError            ErrorCode = "E13009"
	CodeAggregateError         ErrorCode = "E13010"
)

// Cache error codes (14000-14999)
const (
	CodeCacheError             ErrorCode = "E14001"
	CodeCacheMiss              ErrorCode = "E14002"
	CodeCacheTimeout           ErrorCode = "E14003"
	CodeCacheConnectionFailed  ErrorCode = "E14004"
	CodeCacheSerializationError ErrorCode = "E14005"
	CodeCacheDeserializationError ErrorCode = "E14006"
	CodeCacheKeyNotFound       ErrorCode = "E14007"
	CodeCacheKeyExpired        ErrorCode = "E14008"
	CodeCacheMemoryFull        ErrorCode = "E14009"
	CodeCacheEvictionFailed    ErrorCode = "E14010"
)

// Storage error codes (15000-15999)
const (
	CodeStorageError           ErrorCode = "E15001"
	CodeFileNotFound           ErrorCode = "E15002"
	CodeFileAlreadyExists      ErrorCode = "E15003"
	CodeFilePermissionDenied   ErrorCode = "E15004"
	CodeFileCorrupted          ErrorCode = "E15005"
	CodeFileTooLarge           ErrorCode = "E15006"
	CodeDiskFull               ErrorCode = "E15007"
	CodeDirectoryNotFound      ErrorCode = "E15008"
	CodeStorageQuotaExceeded   ErrorCode = "E15009"
	CodeStorageUnavailable     ErrorCode = "E15010"
)

// ErrorCodeInfo provides metadata about error codes
type ErrorCodeInfo struct {
	Code        ErrorCode     `json:"code"`
	Category    ErrorCategory `json:"category"`
	Severity    ErrorSeverity `json:"severity"`
	Recovery    ErrorRecovery `json:"recovery"`
	Description string        `json:"description"`
	HTTPStatus  int           `json:"http_status"`
}

// ErrorCodeRegistry maps error codes to their metadata
var ErrorCodeRegistry = map[ErrorCode]ErrorCodeInfo{
	// Validation errors
	CodeValidationFailed: {
		Code: CodeValidationFailed, Category: CategoryValidation, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Input validation failed", HTTPStatus: 400,
	},
	CodeInvalidInput: {
		Code: CodeInvalidInput, Category: CategoryValidation, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Invalid input provided", HTTPStatus: 400,
	},
	
	// Authentication errors
	CodeAuthenticationFailed: {
		Code: CodeAuthenticationFailed, Category: CategoryAuthentication, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Authentication failed", HTTPStatus: 401,
	},
	CodeTokenExpired: {
		Code: CodeTokenExpired, Category: CategoryAuthentication, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Authentication token expired", HTTPStatus: 401,
	},
	
	// Authorization errors
	CodeAuthorizationFailed: {
		Code: CodeAuthorizationFailed, Category: CategoryAuthorization, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Authorization failed", HTTPStatus: 403,
	},
	CodeRateLimitExceeded: {
		Code: CodeRateLimitExceeded, Category: CategoryRateLimit, Severity: SeverityWarning,
		Recovery: RecoveryRetryable, Description: "Rate limit exceeded", HTTPStatus: 429,
	},
	
	// Resource errors
	CodeResourceNotFound: {
		Code: CodeResourceNotFound, Category: CategoryNotFound, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Resource not found", HTTPStatus: 404,
	},
	CodeResourceConflict: {
		Code: CodeResourceConflict, Category: CategoryConflict, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Resource conflict", HTTPStatus: 409,
	},
	
	// Network errors
	CodeNetworkError: {
		Code: CodeNetworkError, Category: CategoryNetwork, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "Network communication error", HTTPStatus: 502,
	},
	CodeConnectionTimeout: {
		Code: CodeConnectionTimeout, Category: CategoryNetwork, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "Connection timeout", HTTPStatus: 504,
	},
	
	// Database errors
	CodeDatabaseError: {
		Code: CodeDatabaseError, Category: CategoryDatabase, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "Database operation error", HTTPStatus: 500,
	},
	CodeDatabaseTimeout: {
		Code: CodeDatabaseTimeout, Category: CategoryDatabase, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "Database operation timeout", HTTPStatus: 504,
	},
	
	// AI service errors
	CodeAIServiceError: {
		Code: CodeAIServiceError, Category: CategoryAI, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "AI service error", HTTPStatus: 502,
	},
	CodeAIRateLimitExceeded: {
		Code: CodeAIRateLimitExceeded, Category: CategoryAI, Severity: SeverityWarning,
		Recovery: RecoveryRetryable, Description: "AI service rate limit exceeded", HTTPStatus: 429,
	},
	
	// Timeout errors
	CodeTimeout: {
		Code: CodeTimeout, Category: CategoryTimeout, Severity: SeverityError,
		Recovery: RecoveryRetryable, Description: "Operation timeout", HTTPStatus: 504,
	},
	CodeContextCancelled: {
		Code: CodeContextCancelled, Category: CategoryTimeout, Severity: SeverityWarning,
		Recovery: RecoveryNonRetryable, Description: "Context cancelled", HTTPStatus: 499,
	},
	
	// Circuit breaker errors
	CodeCircuitBreakerOpen: {
		Code: CodeCircuitBreakerOpen, Category: CategoryCircuitBreaker, Severity: SeverityWarning,
		Recovery: RecoveryCircuitBreak, Description: "Circuit breaker is open", HTTPStatus: 503,
	},
	
	// Internal errors
	CodeInternalError: {
		Code: CodeInternalError, Category: CategoryInternal, Severity: SeverityCritical,
		Recovery: RecoveryNonRetryable, Description: "Internal system error", HTTPStatus: 500,
	},
}

// GetErrorCodeInfo returns metadata for an error code
func GetErrorCodeInfo(code ErrorCode) (ErrorCodeInfo, bool) {
	info, exists := ErrorCodeRegistry[code]
	return info, exists
}

// IsRetryableErrorCode checks if an error code indicates a retryable error
func IsRetryableErrorCode(code ErrorCode) bool {
	if info, exists := ErrorCodeRegistry[code]; exists {
		return info.Recovery == RecoveryRetryable
	}
	return false
}

// GetHTTPStatusForErrorCode returns the appropriate HTTP status code for an error code
func GetHTTPStatusForErrorCode(code ErrorCode) int {
	if info, exists := ErrorCodeRegistry[code]; exists {
		return info.HTTPStatus
	}
	return 500 // Default to internal server error
}

// GetErrorsByCategory returns all error codes in a specific category
func GetErrorsByCategory(category ErrorCategory) []ErrorCode {
	var codes []ErrorCode
	for code, info := range ErrorCodeRegistry {
		if info.Category == category {
			codes = append(codes, code)
		}
	}
	return codes
}

// GetErrorsBySeverity returns all error codes with a specific severity
func GetErrorsBySeverity(severity ErrorSeverity) []ErrorCode {
	var codes []ErrorCode
	for code, info := range ErrorCodeRegistry {
		if info.Severity == severity {
			codes = append(codes, code)
		}
	}
	return codes
}
