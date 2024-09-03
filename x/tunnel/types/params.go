package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gopkg.in/yaml.v2"
)

var (
	DefaultMinInterval = uint64(1)
	DefaultMinDeposit  = sdk.NewCoins(sdk.NewInt64Coin("band", 1000000))
	DefaultMaxSignals  = uint64(100)
	DefaultBaseFee     = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
)

// NewParams creates a new Params instance
func NewParams(
	minDeposit sdk.Coins,
	minInterval uint64,
	maxSignals uint64,
	baseFee sdk.Coins,
) Params {
	return Params{
		MinDeposit:  minDeposit,
		MinInterval: minInterval,
		MaxSignals:  maxSignals,
		BaseFee:     baseFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinDeposit,
		DefaultMinInterval,
		DefaultMaxSignals,
		DefaultBaseFee,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	// Validate MinDeposit
	if !p.MinDeposit.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf(p.MinDeposit.String())
	}

	// Validate MinInterval
	if err := validateUint64("min interval", true)(p.MinInterval); err != nil {
		return err
	}

	// Validate MaxSignals
	if err := validateUint64("max signals", true)(p.MaxSignals); err != nil {
		return err
	}

	// Validate BaseFee
	if !p.BaseFee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf(p.BaseFee.String())
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateBasisPoint validates if a given number is a valid basis point (0 to 10000).
func validateBasisPoint(name string, bp uint64) error {
	if err := validateUint64(name, false)(bp); err != nil {
		return err
	}

	if bp > 10000 {
		return fmt.Errorf("invalid basis point: must be between 0 and 10000")
	}
	return nil
}

// validateUint64 validates if a given number is a valid uint64.
func validateUint64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
