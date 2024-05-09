package api

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

type SyncCommitteeResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Validators          []string   `json:"validators"`
		ValidatorAggregates [][]string `json:"validator_aggregates"`
	} `json:"data"`
}
