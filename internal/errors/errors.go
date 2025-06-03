package errors

import (
	"errors"
	"fmt"
)

// Re-export standard errors package functions.
var (
	As     = errors.As
	Is     = errors.Is
	New    = errors.New
	Unwrap = errors.Unwrap
)

// Define domain-specific error types.
var (
	// General errors.
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrForbidden        = errors.New("operation not permitted")

	// VM-specific errors.
	ErrVMNotFound           = errors.New("VM not found")
	ErrVMAlreadyExists      = errors.New("VM already exists")
	ErrVMInvalidState       = errors.New("invalid VM state for operation")
	ErrInvalidCPUCount      = errors.New("invalid CPU count")
	ErrInvalidMemorySize    = errors.New("invalid memory size")
	ErrInvalidDiskSize      = errors.New("invalid disk size")
	ErrInvalidDiskFormat    = errors.New("invalid disk format")
	ErrInvalidNetworkType   = errors.New("invalid network type")
	ErrInvalidNetworkSource = errors.New("invalid network source")

	// Storage errors.
	ErrStoragePoolNotFound = errors.New("storage pool not found")
	ErrVolumeNotFound      = errors.New("volume not found")
	ErrInsufficientStorage = errors.New("insufficient storage space")

	// Network errors.
	ErrNetworkNotFound   = errors.New("network not found")
	ErrIPAddressNotFound = errors.New("IP address not found")

	// Authentication errors.
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrDuplicateUsername  = errors.New("username already exists")

	// Export errors.
	ErrExportFailed      = errors.New("export operation failed")
	ErrExportJobNotFound = errors.New("export job not found")
	ErrUnsupportedFormat = errors.New("unsupported export format")
)

// Wrap wraps an error with additional context.
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// WrapWithCode wraps an error with a specific error code.
func WrapWithCode(err error, code error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	wrappedErr := fmt.Errorf(format+": %w", append(args, err)...)
	return fmt.Errorf("%w: %w", code, wrappedErr)
}

// GetErrorCode extracts the error code from an error.
func GetErrorCode(err error) error {
	if err == nil {
		return nil
	}

	// List of all error codes
	errorCodes := []error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidParameter,
		ErrForbidden,
		ErrVMNotFound,
		ErrVMAlreadyExists,
		ErrVMInvalidState,
		ErrInvalidCPUCount,
		ErrInvalidMemorySize,
		ErrInvalidDiskSize,
		ErrInvalidDiskFormat,
		ErrInvalidNetworkType,
		ErrInvalidNetworkSource,
		ErrStoragePoolNotFound,
		ErrVolumeNotFound,
		ErrInsufficientStorage,
		ErrNetworkNotFound,
		ErrIPAddressNotFound,
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrInvalidToken,
		ErrUserInactive,
		ErrDuplicateUsername,
		ErrExportFailed,
		ErrExportJobNotFound,
		ErrUnsupportedFormat,
	}

	// Check if the error is or wraps any of our error codes
	for _, code := range errorCodes {
		if errors.Is(err, code) {
			return code
		}
	}

	return nil
}

// errorCodeStrings maps error codes to their string representations.
var errorCodeStrings = map[error]string{
	ErrNotFound:             "NOT_FOUND",
	ErrAlreadyExists:        "ALREADY_EXISTS",
	ErrInvalidParameter:     "INVALID_PARAMETER",
	ErrForbidden:            "FORBIDDEN",
	ErrVMNotFound:           "VM_NOT_FOUND",
	ErrVMAlreadyExists:      "VM_ALREADY_EXISTS",
	ErrVMInvalidState:       "VM_INVALID_STATE",
	ErrInvalidCPUCount:      "INVALID_CPU_COUNT",
	ErrInvalidMemorySize:    "INVALID_MEMORY_SIZE",
	ErrInvalidDiskSize:      "INVALID_DISK_SIZE",
	ErrInvalidDiskFormat:    "INVALID_DISK_FORMAT",
	ErrInvalidNetworkType:   "INVALID_NETWORK_TYPE",
	ErrInvalidNetworkSource: "INVALID_NETWORK_SOURCE",
	ErrStoragePoolNotFound:  "STORAGE_POOL_NOT_FOUND",
	ErrVolumeNotFound:       "VOLUME_NOT_FOUND",
	ErrInsufficientStorage:  "INSUFFICIENT_STORAGE",
	ErrNetworkNotFound:      "NETWORK_NOT_FOUND",
	ErrIPAddressNotFound:    "IP_ADDRESS_NOT_FOUND",
	ErrInvalidCredentials:   "INVALID_CREDENTIALS",
	ErrTokenExpired:         "TOKEN_EXPIRED",
	ErrInvalidToken:         "INVALID_TOKEN",
	ErrUserInactive:         "USER_INACTIVE",
	ErrDuplicateUsername:    "DUPLICATE_USERNAME",
	ErrExportFailed:         "EXPORT_FAILED",
	ErrExportJobNotFound:    "EXPORT_JOB_NOT_FOUND",
	ErrUnsupportedFormat:    "UNSUPPORTED_FORMAT",
}

// GetErrorCodeString returns the string representation of the error code.
func GetErrorCodeString(err error) string {
	if err == nil {
		return "UNKNOWN_ERROR"
	}

	code := GetErrorCode(err)
	if code == nil {
		return "INTERNAL_SERVER_ERROR"
	}

	if codeString, exists := errorCodeStrings[code]; exists {
		return codeString
	}

	return "INTERNAL_SERVER_ERROR"
}
