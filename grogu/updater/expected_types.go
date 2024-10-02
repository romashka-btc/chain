package updater

import (
	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client"
	rpcclient "github.com/cometbft/cometbft/rpc/client"

	feeds "github.com/bandprotocol/chain/v2/x/feeds/types"
)

type BothanClient interface {
	bothan.Client
}

type FeedQuerier interface {
	QueryCurrentFeeds() (*feeds.QueryCurrentFeedsResponse, error)
	QueryReferenceSourceConfig() (*feeds.QueryReferenceSourceConfigResponse, error)
}

type RemoteClient interface {
	rpcclient.RemoteClient
}
