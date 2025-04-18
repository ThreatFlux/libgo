package errors

import (
	"errors"
	"fmt"
)

// Re-export standard errors package functions
var (
	As     = errors.As
	Is     = errors.Is
	New    = errors.New
	Unwrap = errors.Unwrap
)

// Define domain-specific error types
var (
	// General errors
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrForbidden        = errors.New("operation not permitted")

	// VM-specific errors
	ErrVMNotFound       = errors.New("VM not found")
	ErrVMAlreadyExists  = errors.New("VM already exists")
	ErrVMInvalidState   = errors.New("invalid VM state for operation")
	ErrInvalidCPUCount  = errors.New("invalid CPU count")
	ErrInvalidMemorySize = errors.New("invalid memory size")
	ErrInvalidDiskSize   = errors.New("invalid disk size")
	ErrInvalidDiskFormat = errors.New("invalid disk format")
	ErrInvalidNetworkType = errors.New("invalid network type")
	ErrInvalidNetworkSource = errors.New("invalid network source")

	// Storage errors
	ErrStoragePoolNotFound  = errors.New("storage pool not found")
	ErrVolumeNotFound       = errors.New("volume not found")
	ErrInsufficientStorage  = errors.New("insufficient storage space")

	// Network errors
	ErrNetworkNotFound      = errors.New("network not found")
	ErrIPAddressNotFound    = errors.New("IP address not found")

	// Authentication errors
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidToken         = errors.New("invalid token")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrDuplicateUsername    = errors.New("username already exists")

	// Export errors
	ErrExportFailed         = errors.New("export operation failed")
	ErrExportJobNotFound    = errors.New("export job not found")
	ErrUnsupportedFormat    = errors.New("unsupported export format")
)

// Wrap wraps an error with additional context
func Wrap(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// WrapWithCode wraps an error with a specific error code
func WrapWithCode(err error, code error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	wrappedErr := fmt.Errorf(format+": %w", append(args, err)...)
	return fmt.Errorf("%w: %v", code, wrappedErr)
}

// GetErrorCode extracts the error code from an error
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

// GetErrorCodeString returns the string representation of the error code
func GetErrorCodeString(err error) string {
	code := GetErrorCode(err)
	if code == nil {
		return "UNKNOWN_ERROR"
	}

	switch code {
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrAlreadyExists:
		return "ALREADY_EXISTS"
	case ErrInvalidParameter:
		return "INVALID_PARAMETER"
	case ErrForbidden:
		return "FORBIDDEN"
	case ErrVMNotFound:
		return "VM_NOT_FOUND"
	case ErrVMAlreadyExists:
		return "VM_ALREADY_EXISTS"
	case ErrVMInvalidState:
		return "VM_INVALID_STATE"
	case ErrInvalidCPUCount:
		return "INVALID_CPU_COUNT"
	case ErrInvalidMemorySize:
		return "INVALID_MEMORY_SIZE"
	case ErrInvalidDiskSize:
		return "INVALID_DISK_SIZE"
	case ErrInvalidDiskFormat:
		return "INVALID_DISK_FORMAT"
	case ErrInvalidNetworkType:
		return "INVALID_NETWORK_TYPE"
	case ErrInvalidNetworkSource:
		return "INVALID_NETWORK_SOURCE"
	case ErrStoragePoolNotFound:
		return "STORAGE_POOL_NOT_FOUND"
	case ErrVolumeNotFound:
		return "VOLUME_NOT_FOUND"
	case ErrInsufficientStorage:
		return "INSUFFICIENT_STORAGE"
	case ErrNetworkNotFound:
		return "NETWORK_NOT_FOUND"
	case ErrIPAddressNotFound:
		return "IP_ADDRESS_NOT_FOUND"
	case ErrInvalidCredentials:
		return "INVALID_CREDENTIALS"
	case ErrTokenExpired:
		return "TOKEN_EXPIRED"
	case ErrInvalidToken:
		return "INVALID_TOKEN"
	case ErrUserInactive:
		return "USER_INACTIVE"
	case ErrDuplicateUsername:
		return "DUPLICATE_USERNAME"
	case ErrExportFailed:
		return "EXPORT_FAILED"
	case ErrExportJobNotFound:
		return "EXPORT_JOB_NOT_FOUND"
	case ErrUnsupportedFormat:
		return "UNSUPPORTED_FORMAT"
	default:
		return "INTERNAL_SERVER_ERROR"
	}
}
