package client

import (
	"context"
	"fmt"
	"net/http"
)

// NewFaucetClient creates FaucetClient to create and fund accounts
func NewFaucetClient(endpoint string, client AptosClient) *FaucetClient {
	impl := &FaucetClient{
		APIBase:     APIBase{endpoint: endpoint},
		aptosClient: client,
	}
	return impl
}

// For more details, see https://github.com/aptos-labs/aptos-core/blob/main/crates/aptos-rest-client/src/faucet.rs
type FaucetClient struct {
	APIBase

	aptosClient AptosClient
}

// FundAccount creates an account if it doesn't exist and mints the given amount of coins into account
func (im *FaucetClient) FundAccount(ctx context.Context, address string, amount uint64) error {
	var txHashes []string
	err := request(ctx, http.MethodPost,
		im.APIBase.Endpoint()+"/mint",
		nil, &txHashes, map[string]interface{}{
			"address": address,
			"amount":  amount,
		}, nil)
	if err != nil {
		return err
	}

	if len(txHashes) == 0 {
		return fmt.Errorf("unexpected response")
	}

	for _, txHash := range txHashes {
		if err := im.aptosClient.WaitForTransaction(ctx, txHash); err != nil {
			return err
		}
	}

	return nil
}
