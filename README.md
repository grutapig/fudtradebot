# GRUTA AI trading bot dashboard

Automated trading bot for cryptocurrency futures markets that tests correlation between community sentiment and price movement. Built in 12 days as a proof of concept.

## Concept

This bot validates whether emotional sentiment and FUD (Fear, Uncertainty, Doubt) attacks in crypto communities can predict market behavior. It combines traditional technical analysis with community sentiment data.

## Trading Pairs

The bot monitors three communities and their tokens:
- **GIGGLEUSDT** (Community ID: 1969807538154811438)
- **TOSHIUSDT** (Community ID: 1786006467847368871)  
- **TURTLEUSDT** (Community ID: 1938175945476555178)

## How It Works

### Data Sources

The bot collects data every 60 seconds:
- Price candles from exchange (1-hour for coins, 4-hour for Bitcoin)
- Community sentiment analysis from external Gruta service
- FUD activity levels from external Gruta service

### Analysis

**Technical Analysis:**
- Ichimoku Cloud indicator for both BTC and the trading coin
- Line crossover signals as entry/exit triggers

**Community Sentiment Analysis (via external Gruta service):**
- Community activity trends (sharp rises/drops detection)
- FUD activity monitoring and levels
- Emotional sentiment scoring
- Coordinated FUD attack detection

**Decision Making:**
- Combines Ichimoku signals with sentiment data
- Claude AI validates each trading decision before opening positions
- Claude AI periodically analyzes open positions for closing decisions

### Trading Modes

1. **Normal Mode:** Ichimoku signals filtered by sentiment analysis
2. **FUD Attack Mode:** Automatic SHORT position when coordinated FUD attack is detected

### Position Management
- Fixed position sizes
- No leverage multiplication
- Simple stop-loss/take-profit based on Ichimoku
- All decisions logged for analysis

## GRUTA AI trading bot dashboard

Web interface displays:
- Complete trading history
- Real-time position status
- Decision logs with AI reasoning
- P/L charts and statistics

## Disclaimer

Experimental prototype. Not financial advice. For research purposes only.
