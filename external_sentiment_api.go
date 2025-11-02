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

func FetchExternalSentimentAnalysis(communityID string) (ClaudeSentimentResponse, error) {
	apiKey := os.Getenv(ENV_API_EXTERNAL_SECRET)
	if apiKey == "" {
		return ClaudeSentimentResponse{}, fmt.Errorf("%s not set", ENV_API_EXTERNAL_SECRET)
	}

	baseURL := fmt.Sprintf("https://grutapig.com/grufender/api/external/sentiment/%s", communityID)
	params := url.Values{}
	params.Add("apikey", apiKey)
	params.Add("limit", "50")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Get(fullURL)
	if err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to fetch sentiment analysis: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ClaudeSentimentResponse{}, fmt.Errorf("sentiment API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var sentimentResponse ClaudeSentimentResponse
	if err := json.Unmarshal(body, &sentimentResponse); err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	return sentimentResponse, nil
}
