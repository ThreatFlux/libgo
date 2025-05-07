package vm

import (
	"github.com/threatflux/libgo/internal/errors"
)

// Error mappings from internal/errors
var (
	// VM-specific errors
	ErrVMNotFound       = errors.ErrVMNotFound
	ErrVMAlreadyExists  = errors.ErrVMAlreadyExists
	ErrVMInvalidState   = errors.ErrVMInvalidState
	ErrVMInUse          = errors.ErrForbidden // No direct mapping, using ErrForbidden
	ErrInvalidCPUCount  = errors.ErrInvalidCPUCount
	ErrInvalidMemorySize = errors.ErrInvalidMemorySize
	ErrInvalidDiskSize   = errors.ErrInvalidDiskSize
	ErrInvalidDiskFormat = errors.ErrInvalidDiskFormat
	ErrInvalidNetworkType = errors.ErrInvalidNetworkType
	ErrInvalidNetworkSource = errors.ErrInvalidNetworkSource
)