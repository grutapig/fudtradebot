package main

import (
	"flag"
	"github.com/grutapig/fudtradebot/claude"
	"github.com/joho/godotenv"
	"log"
	"os"
	"sync"
)

var TradingPairs = []TradingPair{
	{
		CommunityID: "1969807538154811438",
		Symbol:      "GIGGLEUSDT",
		Leverage:    1,
		Quantity:    0.2,
	},
	{
		CommunityID: "1786006467847368871",
		Symbol:      "TOSHIUSDT",
		Leverage:    1,
		Quantity:    22000,
	},
	{
		CommunityID: "1938175945476555178",
		Symbol:      "TURTLEUSDT",
		Leverage:    1,
		Quantity:    150,
	},
}

func main() {
	webOnly := flag.Bool("web-only", false, "Start only web server without trading")
	flag.Parse()

	log.Println("Starting trading bot...")
	godotenv.Load()

	if err := InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Backfilling Max/Min P/L for existing positions...")
	if err := BackfillMaxMinPnL(); err != nil {
		log.Printf("Warning: Failed to backfill Max/Min P/L: %v", err)
	} else {
		log.Println("Max/Min P/L backfill completed")
	}
	log.Println("Database initialized successfully")

	go StartWebServer()

	apiKey := os.Getenv(ENV_DEX_KEY)
	secretKey := os.Getenv(ENV_DEX_SECRET)
	grufenderApiURL := os.Getenv(ENV_GRUFENDER_API_URL)
	proxyDSN := os.Getenv(ENV_PROXY_DSN)
	claudeAPIKey := os.Getenv(ENV_CLAUDE_API_KEY)
	claudeMinIntervalMinutes := getEnvAsInt(ENV_CLAUDE_MIN_INTERVAL_MINUTES, 10)

	if apiKey == "" || secretKey == "" {
		log.Fatalf("%s and %s environment variables must be set", ENV_DEX_KEY, ENV_DEX_SECRET)
	}

	if grufenderApiURL == "" {
		log.Fatalf("ENV_GRUFENDER_API_URL is not set.")
	}

	if claudeAPIKey == "" {
		log.Println("Warning: CLAUDE_API_KEY not set, sentiment analysis will be disabled")
	}

	var exchange AsterDexExchange
	var activityClient ExternalActivityClient
	var err error

	if proxyDSN != "" {
		log.Printf("Initializing clients with proxy")

		exchange, err = NewAsterDexExchangeWithProxy(apiKey, secretKey, proxyDSN)
		if err != nil {
			log.Fatalf("Failed to create exchange with proxy: %v", err)
		}

		activityClient, err = NewExternalActivityClientWithProxy(grufenderApiURL, proxyDSN)
		if err != nil {
			log.Fatalf("Failed to create activity client with proxy: %v", err)
		}
	} else {
		exchange = NewAsterDexExchange(apiKey, secretKey)
		activityClient = NewExternalActivityClient(grufenderApiURL)
	}

	var claudeClient *claude.ClaudeApi
	if claudeAPIKey != "" {
		claudeClient, err = claude.NewClaudeClient(claudeAPIKey, proxyDSN, claude.CLAUDE_45_MODEL)
		if err != nil {
			log.Fatalf("Failed to create Claude client: %v", err)
		}
		claudeClient.SetMaxTokens(4000)
	}

	go runBalanceCollector(exchange)

	if *webOnly {
		log.Println("Running in WEB-ONLY mode - trading disabled")
		select {}
	}

	var wg sync.WaitGroup

	for _, pair := range TradingPairs {
		wg.Add(1)
		go func(pair TradingPair) {
			defer wg.Done()
			runTradingLoop(exchange, activityClient, claudeClient, pair, claudeMinIntervalMinutes)
		}(pair)
	}

	wg.Wait()
}
