package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// validateLastSignalPricesList validates the latest signal prices list.
func validateLastSignalPricesList(
	tunnels []types.Tunnel,
	lsps []types.LatestSignalPrices,
) error {
	if len(tunnels) != len(lsps) {
		return fmt.Errorf("tunnels and latest signal prices list length mismatch")
	}

	tunnelMap := make(map[uint64]bool)
	for _, t := range tunnels {
		tunnelMap[t.ID] = true
	}

	for _, lsp := range lsps {
		if _, ok := tunnelMap[lsp.TunnelID]; !ok {
			return fmt.Errorf("tunnel ID %d not found in tunnels", lsp.TunnelID)
		}
		if err := lsp.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data *types.GenesisState) error {
	// validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return types.ErrInvalidGenesis.Wrapf("length of tunnels does not match tunnel count")
	}

	// validate the tunnel IDs
	for _, t := range data.Tunnels {
		if t.ID > data.TunnelCount {
			return types.ErrInvalidGenesis.Wrapf("tunnel count mismatch in tunnels")
		}
	}

	// validate the latest signal prices count
	if len(data.LatestSignalPricesList) != int(data.TunnelCount) {
		return types.ErrInvalidGenesis.Wrapf("length of latest signal prices does not match tunnel count")
	}

	// validate latest signal prices
	if err := validateLastSignalPricesList(data.Tunnels, data.LatestSignalPricesList); err != nil {
		return types.ErrInvalidGenesis.Wrapf("invalid latest signal prices: %s", err.Error())
	}

	// validate the total fees
	if err := data.TotalFees.Validate(); err != nil {
		return types.ErrInvalidGenesis.Wrapf("invalid total fees: %s", err.Error())
	}

	return data.Params.Validate()
}

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *Keeper, data *types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// check if the module account exists
	moduleAcc := k.GetTunnelAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	// set module account if its balance is zero
	if balance := k.GetModuleBalance(ctx); balance.IsZero() {
		k.SetModuleAccount(ctx, moduleAcc)
	}

	// set the tunnel count
	k.SetTunnelCount(ctx, data.TunnelCount)

	// set tunnels
	for _, t := range data.Tunnels {
		k.SetTunnel(ctx, t)
		if t.IsActive {
			k.ActiveTunnelID(ctx, t.ID)
		}
	}

	// set the latest signal prices
	for _, latestSignalPrices := range data.LatestSignalPricesList {
		k.SetLatestSignalPrices(ctx, latestSignalPrices)
	}

	// set the total fees
	k.SetTotalFees(ctx, data.TotalFees)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:                 k.GetParams(ctx),
		TunnelCount:            k.GetTunnelCount(ctx),
		Tunnels:                k.GetTunnels(ctx),
		LatestSignalPricesList: k.GetAllLatestSignalPrices(ctx),
		TotalFees:              k.GetTotalFees(ctx),
	}
}
