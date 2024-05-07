package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

var wg = sync.WaitGroup{}
var m = sync.Mutex{}

func main() {
	rpc := os.Getenv("RPC_ENDPOINT")

	execution_client, err := ethclient.Dial(rpc)
	if err != nil {
		log.Fatalf("Failed to connect to node: %v", err)
	}
	defer execution_client.Close()

	r := gin.Default()
	r.GET("/blockreward/:slot", func(c *gin.Context) {
		slot, ok := new(big.Int).SetString(c.Params.ByName("slot"), 10)
		if !ok {
			log.Fatal("SetString: error")
		}
		url := fmt.Sprintf("%seth/v2/beacon/blocks/%s", rpc, slot)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to fetch beacon block: %v\n", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			getHeaderUrl := fmt.Sprintf("%seth/v1/beacon/headers/%s", rpc, "head")
			resp, err := http.Get(getHeaderUrl)
			if err != nil {
				log.Fatalf("Failed to fetch beacon block: %v\n", err)
			}
			defer resp.Body.Close()
			var body BeaconHeadResponse
			err = json.NewDecoder(resp.Body).Decode(&body)
			if err != nil {
				log.Fatalf("Error decoding response body: %v\n", err)
			}
			currentSlot, _ := strconv.Atoi(body.Data.Header.Message.Slot)
			if currentSlot < int(slot.Int64()) {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "The requested slot is in the future",
				})
				return
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "The slot does not exist / was missed",
				})
				return
			}
		}
		var body GetBlockV2Response
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			log.Fatalf("Error decoding response body: %v\n", err)
		}

		blockNumberString := body.Data.Message.Body.ExecutionPayload.BlockNumber
		blockNumber, ok := new(big.Int).SetString(blockNumberString, 10)
		if !ok {
			log.Fatalf("Error parsing block number: %v\n", blockNumberString)
		}
		block, err := execution_client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			log.Fatalf("Failed to fetch block by number: %v", err)
		}
		baseFeePerGas, _ := strconv.Atoi(body.Data.Message.Body.ExecutionPayload.BaseFeePerGas)
		gasUsed, _ := strconv.Atoi(body.Data.Message.Body.ExecutionPayload.GasUsed)

		baseFee := baseFeePerGas * gasUsed
		var transactionFees uint64
		for _, tx := range block.Transactions() {
			wg.Add(1)
			go addTxFeeToOverallFees(execution_client, tx, &transactionFees)
		}
		wg.Wait()
		c.JSON(http.StatusOK, gin.H{
			"status":           "MEV",
			"transaction_fees": transactionFees,
			"base_fee":         baseFee,
			"block_reward":     transactionFees - uint64(baseFee),
			"block_number":     blockNumberString,
		})

	})

	r.Run(":8080")
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

// ///////////////////// TYPES ///////////////////////////
type GetBlockV2Response struct {
	Version             string    `json:"version"`
	ExecutionOptimistic bool      `json:"execution_optimistic"`
	Finalized           bool      `json:"finalized"`
	Data                BlockData `json:"data"`
}

type BlockData struct {
	Message   BlockMessage `json:"message"`
	Signature string       `json:"signature"`
}

type BlockMessage struct {
	Slot          string    `json:"slot"`
	ProposerIndex string    `json:"proposer_index"`
	ParentRoot    string    `json:"parent_root"`
	StateRoot     string    `json:"state_root"`
	Body          BlockBody `json:"body"`
}

type BlockBody struct {
	RandaoReveal          string           `json:"randao_reveal"`
	Eth1Data              Eth1Data         `json:"eth1_data"`
	Graffiti              string           `json:"graffiti"`
	ProposerSlashings     []interface{}    `json:"proposer_slashings"`
	AttesterSlashings     []interface{}    `json:"attester_slashings"`
	Attestations          []Attestation    `json:"attestations"`
	Deposits              []interface{}    `json:"deposits"`
	VoluntaryExits        []interface{}    `json:"voluntary_exits"`
	SyncAggregate         SyncAggregate    `json:"sync_aggregate"`
	ExecutionPayload      ExecutionPayload `json:"execution_payload"`
	BlsToExecutionChanges []interface{}    `json:"bls_to_execution_changes"`
	BlobKzgCommitments    []interface{}    `json:"blob_kzg_commitments"`
}

type Eth1Data struct {
	DepositRoot  string `json:"deposit_root"`
	DepositCount string `json:"deposit_count"`
	BlockHash    string `json:"block_hash"`
}

type Attestation struct {
	AggregationBits string          `json:"aggregation_bits"`
	Data            AttestationData `json:"data"`
	Signature       string          `json:"signature"`
}

type AttestationData struct {
	Slot            string    `json:"slot"`
	Index           string    `json:"index"`
	BeaconBlockRoot string    `json:"beacon_block_root"`
	Source          EpochRoot `json:"source"`
	Target          EpochRoot `json:"target"`
}

type EpochRoot struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root"`
}

type SyncAggregate struct {
	SyncCommitteeBits      string `json:"sync_committee_bits"`
	SyncCommitteeSignature string `json:"sync_committee_signature"`
}

type ExecutionPayload struct {
	ParentHash    string       `json:"parent_hash"`
	FeeRecipient  string       `json:"fee_recipient"`
	StateRoot     string       `json:"state_root"`
	ReceiptsRoot  string       `json:"receipts_root"`
	LogsBloom     string       `json:"logs_bloom"`
	PrevRandao    string       `json:"prev_randao"`
	BlockNumber   string       `json:"block_number"`
	GasLimit      string       `json:"gas_limit"`
	GasUsed       string       `json:"gas_used"`
	Timestamp     string       `json:"timestamp"`
	ExtraData     string       `json:"extra_data"`
	BaseFeePerGas string       `json:"base_fee_per_gas"`
	BlockHash     string       `json:"block_hash"`
	Transactions  []string     `json:"transactions"`
	Withdrawals   []Withdrawal `json:"withdrawals"`
	BlobGasUsed   string       `json:"blob_gas_used"`
	ExcessBlobGas string       `json:"excess_blob_gas"`
}

type Withdrawal struct {
	Index          string `json:"index"`
	ValidatorIndex string `json:"validator_index"`
	Address        string `json:"address"`
	Amount         string `json:"amount"`
}

type BeaconHeadResponse struct {
	Data struct {
		Header struct {
			Message struct {
				Slot string `json:"slot"`
			} `json:"message"`
		} `json:"header"`
	} `json:"data"`
}
