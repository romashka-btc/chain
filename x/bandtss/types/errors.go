package types

import "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrInvalidAccAddressFormat  = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrInvalidStatus            = errors.Register(ModuleName, 2, "invalid status")
	ErrStatusIsNotActive        = errors.Register(ModuleName, 3, "status is not active")
	ErrTooSoonToActivate        = errors.Register(ModuleName, 4, "too soon to activate")
	ErrRequestReplacementFailed = errors.Register(ModuleName, 5, "failed to request replacement")
	ErrNotEnoughFee             = errors.Register(ModuleName, 6, "not enough fee")
)
