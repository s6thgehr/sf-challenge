package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/s6thgehr/sf-challenge/api"
)

var wg = sync.WaitGroup{}
var m = sync.Mutex{}

func BlockRewardHandler(client *ethclient.Client, rpc string) gin.HandlerFunc {
	return func(c *gin.Context) {
		slot, err := strconv.Atoi(c.Param("slot"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		beaconBlock, err := api.FetchBeaconBlockBySlot(rpc, slot)
		if err != nil {
			currentSlot, err := api.FetchCurrentSlot(rpc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
				return
			}

			if *currentSlot < slot {
				c.JSON(http.StatusBadRequest, gin.H{"message": "requested slot is in the future"})
				return
			} else {
				c.JSON(http.StatusNotFound, gin.H{"message": "slot does not exist / was missed"})
				return
			}
		}

		executionPayload := beaconBlock.Data.Message.Body.ExecutionPayload

		baseFeePerGas, err := strconv.Atoi(executionPayload.BaseFeePerGas)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		gasUsed, err := strconv.Atoi(executionPayload.GasUsed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		blockNumber, err := strconv.Atoi(executionPayload.BlockNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		block, err := api.FetchBlockByNumber(client, blockNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		}

		feeRecipient := strings.ToLower(executionPayload.FeeRecipient)
		baseFee := baseFeePerGas * gasUsed
		status := VANILLA_BLOCK

		var transactionFees uint64
		var reward uint64
		var statusText string
		for _, tx := range block.Transactions() {
			senderAddress, err := getFrom(tx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
				return
			}
			if senderAddress == feeRecipient {
				status = MEV_RELAY
				reward = tx.Value().Uint64()
				statusText = "produced by a MEV relay"
			}
			wg.Add(1)
			go addTxFeeToOverallFees(client, tx, &transactionFees)
		}
		wg.Wait()

		if status == VANILLA_BLOCK {
			reward = transactionFees - uint64(baseFee)
			statusText = "built internally in the validator node"
		}

		c.JSON(http.StatusOK, gin.H{
			"reward": float64(reward) / 1e9,
			"status": statusText,
		})
	}
}

func addTxFeeToOverallFees(client *ethclient.Client, tx *types.Transaction, transactionFees *uint64) error {
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to fetch tx receipt: %v", err)
	}

	m.Lock()
	*transactionFees += receipt.GasUsed * receipt.EffectiveGasPrice.Uint64()
	m.Unlock()
	wg.Done()

	return nil
}

func getFrom(tx *types.Transaction) (string, error) {
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	return strings.ToLower(from.String()), err
}

type Status int

const (
	VANILLA_BLOCK Status = iota
	MEV_RELAY
)
