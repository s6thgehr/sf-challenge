package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

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

func ConvertValidatorIndexToPubkey(rpc string, validatorInds []string, slot int, channel chan string) error {
	defer close(channel)

	currentSlot, err := FetchCurrentSlot(rpc)
	if err != nil {
		return fmt.Errorf("failed to fetch current slot: %v", err)
	}

	var url string
	if *currentSlot < slot {
		url = fmt.Sprintf("%seth/v1/beacon/states/head/validators", rpc)
	} else {
		url = fmt.Sprintf("%seth/v1/beacon/states/%d/validators", rpc, slot)
	}

	for _, validatorIndex := range validatorInds {
		wg.Add(1)
		time.Sleep(10 * time.Millisecond)
		go fetchValidorAddressFromIndex(fmt.Sprintf("%s/%s", url, validatorIndex), channel)
	}
	wg.Wait()
	return nil
}

func fetchValidorAddressFromIndex(url string, channel chan string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch validator info: %v", err)
	}
	defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("failed to fetch validator pubkey: %d", resp.StatusCode)
	// }

	var vr ValidatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return fmt.Errorf("failed to decode validator response: %v", err)
	}

	channel <- vr.Data.Validator.PubKey
	wg.Done()
	return nil
}
