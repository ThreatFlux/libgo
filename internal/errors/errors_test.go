package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "context")
	
	if wrappedErr == nil {
		t.Fatal("Wrap() returned nil for non-nil error")
	}
	
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("Wrap() did not preserve original error for error checking")
	}
	
	expectedMsg := "context: original error"
	if wrappedErr.Error() != expectedMsg {
		t.Errorf("Wrap() produced unexpected message: got %q, want %q", wrappedErr.Error(), expectedMsg)
	}
	
	// Test with formatting
	formattedErr := Wrap(originalErr, "context with %s", "format")
	expectedFormattedMsg := "context with format: original error"
	if formattedErr.Error() != expectedFormattedMsg {
		t.Errorf("Wrap() with format produced unexpected message: got %q, want %q",
			formattedErr.Error(), expectedFormattedMsg)
	}
	
	// Test with nil error
	if nilErr := Wrap(nil, "context"); nilErr != nil {
		t.Errorf("Wrap(nil, ...) should return nil, got %v", nilErr)
	}
}

func TestWrapWithCode(t *testing.T) {
	originalErr := errors.New("original error")
	codedErr := WrapWithCode(originalErr, ErrNotFound, "context")
	
	if codedErr == nil {
		t.Fatal("WrapWithCode() returned nil for non-nil error")
	}
	
	// Error code test
	if !errors.Is(codedErr, ErrNotFound) {
		t.Errorf("WrapWithCode() did not preserve error code for error checking")
	}
	
	// Original error test
	if !errors.Is(codedErr, originalErr) {
		t.Errorf("WrapWithCode() did not preserve original error for error checking")
	}
	
	// Test with formatting
	formattedErr := WrapWithCode(originalErr, ErrForbidden, "context with %s", "format")
	if !errors.Is(formattedErr, ErrForbidden) {
		t.Errorf("WrapWithCode() with format did not preserve error code")
	}
	
	// Test with nil error
	if nilErr := WrapWithCode(nil, ErrNotFound, "context"); nilErr != nil {
		t.Errorf("WrapWithCode(nil, ...) should return nil, got %v", nilErr)
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "direct error code",
			err:      ErrNotFound,
			expected: ErrNotFound,
		},
		{
			name:     "wrapped error code",
			err:      fmt.Errorf("context: %w", ErrVMNotFound),
			expected: ErrVMNotFound,
		},
		{
			name:     "double wrapped error code",
			err:      fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", ErrInvalidParameter)),
			expected: ErrInvalidParameter,
		},
		{
			name:     "error with no code",
			err:      errors.New("some random error"),
			expected: nil,
		},
		{
			name:     "WrapWithCode result",
			err:      WrapWithCode(errors.New("original"), ErrForbidden, "context"),
			expected: ErrForbidden,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code := GetErrorCode(tc.err)
			if code != tc.expected {
				t.Errorf("GetErrorCode() = %v, want %v", code, tc.expected)
			}
		})
	}
}

func TestGetErrorCodeString(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "UNKNOWN_ERROR",
		},
		{
			name:     "not found error",
			err:      ErrNotFound,
			expected: "NOT_FOUND",
		},
		{
			name:     "VM not found error",
			err:      ErrVMNotFound,
			expected: "VM_NOT_FOUND",
		},
		{
			name:     "wrapped VM not found error",
			err:      fmt.Errorf("context: %w", ErrVMNotFound),
			expected: "VM_NOT_FOUND",
		},
		{
			name:     "error with no code",
			err:      errors.New("some random error"),
			expected: "INTERNAL_SERVER_ERROR",
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			codeStr := GetErrorCodeString(tc.err)
			if codeStr != tc.expected {
				t.Errorf("GetErrorCodeString() = %q, want %q", codeStr, tc.expected)
			}
		})
	}
}

func TestErrorCodesAreUnique(t *testing.T) {
	// This test ensures all error codes have unique error messages
	// which is important for error code comparison
	
	errorCodes := []error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidParameter,
		ErrForbidden,
		ErrVMNotFound,
		ErrVMAlreadyExists,
		ErrVMInvalidState,
		ErrStoragePoolNotFound,
		ErrVolumeNotFound,
		ErrInsufficientStorage,
		ErrNetworkNotFound,
		ErrIPAddressNotFound,
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrInvalidToken,
		ErrExportFailed,
		ErrExportJobNotFound,
		ErrUnsupportedFormat,
	}
	
	seen := make(map[string]error)
	for _, code := range errorCodes {
		msg := code.Error()
		if existing, found := seen[msg]; found {
			t.Errorf("Duplicate error message %q in error codes %#v and %#v", 
				msg, existing, code)
		}
		seen[msg] = code
	}
}

func TestErrorsPackageIntegration(t *testing.T) {
	// Test that our package properly re-exports the standard errors package
	originalErr := errors.New("standard error")
	ourErr := New("our error")
	
	// Test Is
	wrappedErr := fmt.Errorf("wrapped: %w", ourErr)
	if !Is(wrappedErr, ourErr) {
		t.Errorf("Our Is() function does not work properly")
	}
	
	// Test As
	var err error
	if !As(wrappedErr, &err) {
		t.Errorf("Our As() function does not work properly")
	}
	
	// Test Unwrap
	unwrapped := Unwrap(wrappedErr)
	if unwrapped != ourErr {
		t.Errorf("Our Unwrap() function does not work properly")
	}
	
	// Test standard errors integration
	stdWrapped := fmt.Errorf("std wrapped: %w", originalErr)
	if !errors.Is(stdWrapped, originalErr) {
		t.Errorf("Standard errors.Is and our package don't interoperate")
	}
	
	// Test that our error types work with standard errors
	stdWrappedDomain := fmt.Errorf("domain wrapped: %w", ErrNotFound) 
	if !errors.Is(stdWrappedDomain, ErrNotFound) {
		t.Errorf("Our domain errors don't work with standard errors.Is")
	}
}
