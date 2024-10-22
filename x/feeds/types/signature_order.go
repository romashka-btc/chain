package types

import tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"

// signature order types
const (
	SignatureOrderTypeFeeds = "Feeds"
)

// Implements Content Interface
var _ tsstypes.Content = &FeedsSignatureOrder{}

// NewFeedSignatureOrder returns a new FeedSignatureOrder object
func NewFeedSignatureOrder(signalIDs []string, encoder Encoder) *FeedsSignatureOrder {
	return &FeedsSignatureOrder{signalIDs, encoder}
}

// OrderRoute returns the order router key
func (f *FeedsSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "Feeds"
func (f *FeedsSignatureOrder) OrderType() string {
	return SignatureOrderTypeFeeds
}

// IsInternal returns false for FeedsSignatureOrder (allow user to submit this content type).
func (f *FeedsSignatureOrder) IsInternal() bool { return false }

// ValidateBasic validates the request's title and description of the request signature
func (f *FeedsSignatureOrder) ValidateBasic() error {
	if len(f.SignalIDs) == 0 {
		return ErrInvalidSignalIDs
	}

	if f.Encoder == ENCODER_UNSPECIFIED {
		return ErrInvalidEncoder
	}

	return nil
}
