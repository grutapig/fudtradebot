package main

import (
	"encoding/json"
	"fmt"
	"github.com/grutapig/fudtradebot/claude"
)

func AnalyzeFudAttack(claudeClient claude.ClaudeApi, tweets []CommunityTweet) (ClaudeFudAttackResponse, error) {
	systemPrompt := `You are an expert at detecting coordinated FUD (Fear, Uncertainty, Doubt) attacks in cryptocurrency communities.

Your task is to analyze the last 200 messages and detect if there are signs of a coordinated mass FUD attack aimed at sharply dropping the price.

WHAT TO LOOK FOR:
1. Multiple participants supporting each other's negative narratives
2. Coordinated timing of negative messages
3. Similar talking points or themes across multiple users
4. Conversations between users reinforcing FUD
5. Sudden increase in negative sentiment that seems artificial

FUD TYPES:
- "technical" - concerns about project technology, bugs, exploits
- "team" - doubts about team competence, scam accusations
- "market" - price manipulation concerns, whale dumps
- "competitor" - comparisons with competitors showing this project negatively
- "regulatory" - legal or regulatory concerns
- "general" - general negativity without specific theme

IMPORTANT:
- Only flag as attack if there's clear coordination (2+ users reinforcing each other)
- Single negative messages are NOT attacks
- Legitimate concerns are NOT attacks
- Must see dialogue/interaction between participants

Response MUST be in JSON format:
{
  "has_attack": true,
  "confidence": 0.85,
  "message_count": 15,
  "participants": [
    {"username": "user1", "message_count": 7},
    {"username": "user2", "message_count": 8}
  ],
  "fud_type": "technical",
  "theme": "Claims of critical bug in smart contract with users supporting each other",
  "started_hours_ago": 2,
  "justification": "Users user1 and user2 engaged in coordinated discussion about alleged contract vulnerability, reinforcing each other's concerns and creating panic. Pattern shows artificial timing and similar language."
}

If NO attack detected, return:
{
  "has_attack": false,
  "confidence": 0.9,
  "message_count": 0,
  "participants": [],
  "fud_type": "",
  "theme": "",
  "started_hours_ago": 0,
  "justification": "No signs of coordinated FUD attack. Community discussions appear organic."
}`

	tweetsJSON, err := json.Marshal(tweets)
	if err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("failed to marshal tweets: %w", err)
	}

	userPrompt := fmt.Sprintf("Analyze these %d recent messages for coordinated FUD attacks:\n\n%s", len(tweets), string(tweetsJSON))

	response, err := claudeClient.GetCompletion(systemPrompt, userPrompt)
	if err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("claude API error: %w", err)
	}

	var result ClaudeFudAttackResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return ClaudeFudAttackResponse{}, fmt.Errorf("failed to parse response: %w, response: %s", err, response)
	}

	return result, nil
}
