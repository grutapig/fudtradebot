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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClaudeSentimentResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ClaudeSentimentResponse{
			OverallSentiment: 5,
			SentimentTrend:   "neutral",
			Confidence:       0.0,
			KeyThemes:        []string{},
			Recommendation:   fmt.Sprintf("Sentiment API error (status %d): external service unavailable", resp.StatusCode),
		}, nil
	}

	var sentimentResponse ClaudeSentimentResponse
	if err := json.Unmarshal(body, &sentimentResponse); err != nil {
		return ClaudeSentimentResponse{
			OverallSentiment: 5,
			SentimentTrend:   "neutral",
			Confidence:       0.0,
			KeyThemes:        []string{},
			Recommendation:   fmt.Sprintf("Sentiment API parse error: %v", err),
		}, nil
	}

	return sentimentResponse, nil
}
