package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func FetchBeaconBlockBySlot(rpc string, slot int) (*GetBlockV2Response, error) {
	url := fmt.Sprintf("%seth/v2/beacon/blocks/%d", rpc, slot)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch beacon block: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("beacon block not found, status code: %d", resp.StatusCode)
	}

	var response GetBlockV2Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response body: %v", err)
	}
	return &response, nil
}

func FetchCurrentSlot(rpc string) (*int, error) {
	url := fmt.Sprintf("%seth/v1/beacon/headers/%s", rpc, "head")
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch beacon block: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch beacon block: %d", resp.StatusCode)
	}

	var body BeaconHeadResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	slot, err := strconv.Atoi(body.Data.Header.Message.Slot)
	if err != nil {
		return nil, fmt.Errorf("failed to parse slot: %v", err)
	}

	return &slot, nil
}

func FetchSyncCommittee(rpc string, slot int) ([]string, error) {
	url := fmt.Sprintf("%seth/v1/beacon/states/%d/sync_committees", rpc, slot)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sync committee: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch sync committee: %d", resp.StatusCode)
	}

	var body SyncCommitteeResponse
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return body.Data.Validators, nil
}
