package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type ActivityDataPoint struct {
	Timestamp    int64 `json:"timestamp"`
	MessageCount int   `json:"message_count"`
}

type ActivityResponse struct {
	Status  string              `json:"status"`
	Data    []ActivityDataPoint `json:"data,omitempty"`
	Message string              `json:"message,omitempty"`
	Error   string              `json:"error,omitempty"`
}

type ExternalActivityClient struct {
	baseURL string
	client  http.Client
}

func NewExternalActivityClient(baseURL string) ExternalActivityClient {
	return ExternalActivityClient{
		baseURL: baseURL,
		client: http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewExternalActivityClientWithProxy(baseURL string, proxyDSN string) (ExternalActivityClient, error) {
	transport := &http.Transport{}
	if proxyDSN != "" {
		proxyURL, err := url.Parse(proxyDSN)
		if err != nil {
			return ExternalActivityClient{}, fmt.Errorf("new activity client proxy dsn error: %s", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return ExternalActivityClient{
		baseURL: baseURL,
		client: http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}, nil
}

func (c ExternalActivityClient) GetCommunityActivity(communityID string, timestampFrom, timestampTo int64, period string) ([]ActivityDataPoint, error) {
	endpoint := fmt.Sprintf("%s/api/external/community/%s/activity", c.baseURL, communityID)
	return c.fetchActivity(endpoint, timestampFrom, timestampTo, period)
}

func (c ExternalActivityClient) GetCommunityFudActivity(communityID string, timestampFrom, timestampTo int64, period string) ([]ActivityDataPoint, error) {
	endpoint := fmt.Sprintf("%s/api/external/community/%s/fud-activity", c.baseURL, communityID)
	return c.fetchActivity(endpoint, timestampFrom, timestampTo, period)
}

func (c ExternalActivityClient) fetchActivity(endpoint string, timestampFrom, timestampTo int64, period string) ([]ActivityDataPoint, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	if timestampFrom > 0 {
		query.Set("timestamp_from", strconv.FormatInt(timestampFrom, 10))
	}
	if timestampTo > 0 {
		query.Set("timestamp_to", strconv.FormatInt(timestampTo, 10))
	}
	if period != "" {
		query.Set("period", period)
	}
	u.RawQuery = query.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ActivityResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status == "error" {
		return nil, fmt.Errorf("%s: %s", result.Message, result.Error)
	}

	return result.Data, nil
}

type ActivityTrend string

const (
	ActivityTrendSharpRise ActivityTrend = "SHARP_RISE"
	ActivityTrendSharpDrop ActivityTrend = "SHARP_DROP"
	ActivityTrendPlateau   ActivityTrend = "PLATEAU"
)

type ActivityAnalysis struct {
	Trend              ActivityTrend
	AverageCount       float64
	RecentAverageCount float64
	ChangePercent      float64
}

func AnalyzeActivityTrend(data []ActivityDataPoint) ActivityAnalysis {
	if len(data) == 0 {
		return ActivityAnalysis{
			Trend: ActivityTrendPlateau,
		}
	}

	if len(data) == 1 {
		return ActivityAnalysis{
			Trend:              ActivityTrendPlateau,
			AverageCount:       float64(data[0].MessageCount),
			RecentAverageCount: float64(data[0].MessageCount),
			ChangePercent:      0,
		}
	}

	splitPoint := len(data) / 2
	if splitPoint == 0 {
		splitPoint = 1
	}

	oldPeriod := data[:splitPoint]
	recentPeriod := data[splitPoint:]

	oldAvg := calculateAverage(oldPeriod)
	recentAvg := calculateAverage(recentPeriod)

	var changePercent float64
	if oldAvg > 0 {
		changePercent = ((recentAvg - oldAvg) / oldAvg) * 100
	}

	var trend ActivityTrend
	if changePercent >= 50 {
		trend = ActivityTrendSharpRise
	} else if changePercent <= -50 {
		trend = ActivityTrendSharpDrop
	} else {
		trend = ActivityTrendPlateau
	}

	return ActivityAnalysis{
		Trend:              trend,
		AverageCount:       oldAvg,
		RecentAverageCount: recentAvg,
		ChangePercent:      changePercent,
	}
}

func AnalyzeFudActivityTrend(data []ActivityDataPoint) ActivityAnalysis {
	return AnalyzeActivityTrend(data)
}

func calculateAverage(data []ActivityDataPoint) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0
	for _, point := range data {
		sum += point.MessageCount
	}

	return float64(sum) / float64(len(data))
}

func (c ExternalActivityClient) GetRecentTweets(communityID string, limit int) ([]CommunityTweet, error) {
	endpoint := fmt.Sprintf("%s/api/external/community/%s/tweets", c.baseURL, communityID)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	u.RawQuery = query.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TweetsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status == "error" {
		return nil, fmt.Errorf("%s: %s", result.Message, result.Error)
	}

	return result.Data, nil
}
