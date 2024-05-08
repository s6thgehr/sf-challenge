package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		beaconBlock, err := api.FetchBeaconBlock(rpc, slot)
		if err != nil {
			currentSlot, err := api.FetchCurrentSlot(rpc)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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

		baseFeePerGas, err := strconv.Atoi(beaconBlock.Data.Message.Body.ExecutionPayload.BaseFeePerGas)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		gasUsed, err := strconv.Atoi(beaconBlock.Data.Message.Body.ExecutionPayload.GasUsed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		baseFee := baseFeePerGas * gasUsed

		blockNumber, err := strconv.Atoi(beaconBlock.Data.Message.Body.ExecutionPayload.BlockNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		block, err := api.FetchBlockByNumber(client, int64(blockNumber))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		var transactionFees uint64

		for _, tx := range block.Transactions() {
			wg.Add(1)
			go addTxFeeToOverallFees(client, tx, &transactionFees)
		}
		wg.Wait()

		c.JSON(http.StatusOK, gin.H{
			"status":           "MEV",
			"transaction_fees": transactionFees,
			"base_fee":         baseFee,
			"block_reward":     transactionFees - uint64(baseFee),
			"block_number":     blockNumber,
		})
	}
}

func addTxFeeToOverallFees(client *ethclient.Client, tx *types.Transaction, transactionFees *uint64) {
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Fatalf("Failed to fetch tx receipt: %v", err)
	}
	m.Lock()
	*transactionFees += receipt.GasUsed * receipt.EffectiveGasPrice.Uint64()
	m.Unlock()
	wg.Done()
}
