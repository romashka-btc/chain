package types

// restake module event types
const (
	EventTypeLockPower       = "lock_power"
	EventTypeCreateVault     = "create_vault"
	EventTypeDeactivateVault = "deactivate_vault"
	EventTypeStake           = "stake"
	EventTypeUnstake         = "unstake"

	AttributeKeyStaker = "staker"
	AttributeKeyKey    = "key"
	AttributeKeyPower  = "power"
	AttributeKeyCoins  = "coins"
)
