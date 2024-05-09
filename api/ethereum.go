package api

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func NewEthereumClient(rpc string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %v", err)
	}
	return client, nil
}

func FetchBlockByNumber(client *ethclient.Client, number int) (*types.Block, error) {
	return client.BlockByNumber(context.Background(), big.NewInt(int64(number)))
}
