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

	var resp *http.Response
	var err error

	for attempt := 0; attempt < 2; attempt++ {
		resp, err = client.Get(fullURL)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if attempt == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("failed to fetch FUD attack analysis after 2 attempts: %w", err)
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
			Justification: fmt.Sprintf("FUD API error (status %d): external service unavailable, raw: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("FUD API error (status %d): external service unavailable, raw: %s", resp.StatusCode, string(body))
	}

	var fudResponse ClaudeFudAttackResponse
	if err := json.Unmarshal(body, &fudResponse); err != nil {
		return ClaudeFudAttackResponse{
			HasAttack:     false,
			Confidence:    0.0,
			MessageCount:  0,
			Participants:  []FudAttackParticipant{},
			Justification: fmt.Sprintf("FUD API parse error: %v", err),
		}, fmt.Errorf("FUD API parse error: %v", err)
	}

	return fudResponse, nil
}
