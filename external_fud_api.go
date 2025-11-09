package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

func FetchExternalFudAttackAnalysis(communityID string) (ClaudeFudAttackResponse, error) {
	apiKey := os.Getenv(ENV_API_EXTERNAL_SECRET)
	if apiKey == "" {
		return ClaudeFudAttackResponse{}, fmt.Errorf("%s not set", ENV_API_EXTERNAL_SECRET)
	}

	baseURL := fmt.Sprintf("https://grutapig.com/grufender/api/external/fud-alert/%s", communityID)
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("limit", "200")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Get(fullURL)
	if err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("failed to fetch FUD attack analysis: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ClaudeFudAttackResponse{
			HasAttack:     false,
			Confidence:    0.0,
			MessageCount:  0,
			Participants:  []FudAttackParticipant{},
			Justification: fmt.Sprintf("FUD API error (status %d): external service unavailable", resp.StatusCode),
		}, nil
	}

	var fudResponse ClaudeFudAttackResponse
	if err := json.Unmarshal(body, &fudResponse); err != nil {
		return ClaudeFudAttackResponse{
			HasAttack:     false,
			Confidence:    0.0,
			MessageCount:  0,
			Participants:  []FudAttackParticipant{},
			Justification: fmt.Sprintf("FUD API parse error: %v", err),
		}, nil
	}

	return fudResponse, nil
}
