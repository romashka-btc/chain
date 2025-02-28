package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// SetNewGroupTransition sets a new group transition with the current group ID and public key.
func (k Keeper) SetNewGroupTransition(
	ctx sdk.Context,
	incomingGroupID tss.GroupID,
	execTime time.Time,
	isForceTransition bool,
) (types.GroupTransition, error) {
	status := types.TRANSITION_STATUS_CREATING_GROUP
	if isForceTransition {
		status = types.TRANSITION_STATUS_WAITING_EXECUTION
	}

	// get the current group ID and public key.
	currentGroupID := k.GetCurrentGroup(ctx).GroupID
	var currentGroupPubKey tss.Point
	if currentGroupID != 0 {
		currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
		if err != nil {
			return types.GroupTransition{}, err
		}
		currentGroupPubKey = currentGroup.PubKey
	}

	// get incoming group and its public key.
	var incomingGroupPubKey tss.Point
	if isForceTransition {
		incomingGroup, err := k.tssKeeper.GetGroup(ctx, incomingGroupID)
		if err != nil {
			return types.GroupTransition{}, err
		}
		incomingGroupPubKey = incomingGroup.PubKey
	}

	transition := types.NewGroupTransition(
		tss.SigningID(0),
		currentGroupID,
		incomingGroupID,
		currentGroupPubKey,
		incomingGroupPubKey,
		status,
		execTime,
		isForceTransition,
	)
	k.SetGroupTransition(ctx, transition)

	return transition, nil
}

// EndGroupTransitionProcess ends the group transition process by removing the transition and emit
// an event.
func (k Keeper) EndGroupTransitionProcess(ctx sdk.Context, transition types.GroupTransition, isSuccess bool) {
	eventType := types.EventTypeGroupTransitionSuccess
	if !isSuccess {
		eventType = types.EventTypeGroupTransitionFailed
	}

	k.DeleteGroupTransition(ctx)
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		eventType,
		sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", transition.SigningID)),
		sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", transition.CurrentGroupID)),
		sdk.NewAttribute(types.AttributeKeyIncomingGroupID, fmt.Sprintf("%d", transition.IncomingGroupID)),
	))
}

// ShouldExecuteGroupTransition checks if the group transition should be executed by comparing
// the block time with the transition execution time.
func (k Keeper) ShouldExecuteGroupTransition(ctx sdk.Context) (transition types.GroupTransition, ok bool) {
	transition, found := k.GetGroupTransition(ctx)
	if !found || transition.ExecTime.After(ctx.BlockTime()) {
		return types.GroupTransition{}, false
	}

	return transition, true
}

// ExecuteGroupTransition executes the group transition by updating the current group ID and
// removing the members in the previous group if the status is waiting execution, or ends
// the process otherwise.
func (k Keeper) ExecuteGroupTransition(ctx sdk.Context, transition types.GroupTransition) {
	if transition.Status != types.TRANSITION_STATUS_WAITING_EXECUTION {
		k.EndGroupTransitionProcess(ctx, transition, false)
		return
	}

	// update current group and delete members in previous group if the status is waiting transition.
	if transition.CurrentGroupID != 0 {
		k.DeleteMembers(ctx, transition.CurrentGroupID)
	}

	newCurrentGroup := types.NewCurrentGroup(transition.IncomingGroupID, transition.ExecTime)
	k.SetCurrentGroup(ctx, newCurrentGroup)
	k.EndGroupTransitionProcess(ctx, transition, true)
}

// ValidateTransitionExecTime validate the transition execution time if it is
// after the block time but not over the max duration.
func (k Keeper) ValidateTransitionExecTime(ctx sdk.Context, execTime time.Time) error {
	params := k.GetParams(ctx)
	minExecTime := ctx.BlockTime().Add(params.MinTransitionDuration)
	maxExecTime := ctx.BlockTime().Add(params.MaxTransitionDuration)

	if execTime.Before(minExecTime) || execTime.After(maxExecTime) {
		return types.ErrInvalidExecTime.Wrapf("exec time should be between %s and %s", minExecTime, maxExecTime)
	}

	return nil
}

// ValidateTransitionInProgress checks if there is a group transition in progress.
func (k Keeper) ValidateTransitionInProgress(ctx sdk.Context) error {
	if _, found := k.GetGroupTransition(ctx); found {
		return types.ErrTransitionInProgress
	}

	return nil
}

// GetIncomingGroupID returns the incoming group ID from transition state. If the status is not
// WaitingExecution, it returns 0.
func (k Keeper) GetIncomingGroupID(ctx sdk.Context) tss.GroupID {
	transition, found := k.GetGroupTransition(ctx)
	if found && transition.Status == types.TRANSITION_STATUS_WAITING_EXECUTION {
		return transition.IncomingGroupID
	}

	return 0
}

// ExtractEventAttributesFromTransition returns the attributes for the group transition event.
func (k Keeper) ExtractEventAttributesFromTransition(transition types.GroupTransition) []sdk.Attribute {
	return []sdk.Attribute{
		sdk.NewAttribute(tsstypes.AttributeKeySigningID, fmt.Sprintf("%d", transition.SigningID)),
		sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", transition.CurrentGroupID)),
		sdk.NewAttribute(types.AttributeKeyCurrentGroupPubKey, transition.CurrentGroupPubKey.String()),
		sdk.NewAttribute(types.AttributeKeyIncomingGroupID, fmt.Sprintf("%d", transition.IncomingGroupID)),
		sdk.NewAttribute(types.AttributeKeyIncomingGroupPubKey, transition.IncomingGroupPubKey.String()),
		sdk.NewAttribute(types.AttributeKeyTransitionStatus, transition.Status.String()),
		sdk.NewAttribute(types.AttributeKeyExecTime, transition.ExecTime.Format(time.RFC3339)),
	}
}

// CreateTransitionSigning creates a signing request for the group transition.
func (k Keeper) CreateTransitionSigning(
	ctx sdk.Context,
	groupPubKey tss.Point,
	transitionTime time.Time,
) (tss.SigningID, error) {
	currentGroupID := k.GetCurrentGroup(ctx).GroupID

	moduleAcc := k.GetBandtssAccount(ctx)
	originator := tsstypes.NewDirectOriginator(ctx.ChainID(), moduleAcc.GetAddress().String(), "")

	content := types.NewGroupTransitionSignatureOrder(groupPubKey, transitionTime)

	signingID, err := k.tssKeeper.RequestSigning(ctx, currentGroupID, &originator, content)
	if err != nil {
		return 0, err
	}

	return signingID, nil
}

// =====================================
// Transition store
// =====================================

// SetGroupTransition sets a group transition information in the store.
func (k Keeper) SetGroupTransition(ctx sdk.Context, groupTransition types.GroupTransition) {
	ctx.KVStore(k.storeKey).Set(types.GroupTransitionStoreKey, k.cdc.MustMarshal(&groupTransition))
}

// GetGroupTransition retrieves a group transition information in the store.
func (k Keeper) GetGroupTransition(ctx sdk.Context) (groupTransition types.GroupTransition, found bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupTransitionStoreKey)
	if bz == nil {
		return groupTransition, false
	}

	k.cdc.MustUnmarshal(bz, &groupTransition)
	return groupTransition, true
}

// DeleteGroupTransition removes the group transition information from the store.
func (k Keeper) DeleteGroupTransition(ctx sdk.Context) {
	ctx.KVStore(k.storeKey).Delete(types.GroupTransitionStoreKey)
}
