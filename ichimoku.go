package main

import (
	"math"
	"strconv"
)

type IchimokuSignal string

const (
	IchimokuSignalStrongLong  IchimokuSignal = "STRONG_LONG"
	IchimokuSignalLong        IchimokuSignal = "LONG"
	IchimokuSignalStrongShort IchimokuSignal = "STRONG_SHORT"
	IchimokuSignalShort       IchimokuSignal = "SHORT"
	IchimokuSignalNeutral     IchimokuSignal = "NEUTRAL"
	IchimokuSignalUncertain   IchimokuSignal = "UNCERTAIN"
)

type IchimokuLine struct {
	Timestamp int64
	Value     float64
}

type IchimokuData struct {
	Tenkan  []IchimokuLine
	Kijun   []IchimokuLine
	SenkouA []IchimokuLine
	SenkouB []IchimokuLine
	Chikou  []IchimokuLine
	Price   []IchimokuLine
}

type IchimokuAnalysis struct {
	Signal             IchimokuSignal
	PriceAboveCloud    bool
	PriceBelowCloud    bool
	PriceInCloud       bool
	BullishCloud       bool
	TenkanAboveKijun   bool
	TwoCloseAboveCloud bool
	TwoCloseBelowCloud bool
	CloudBreakoutUp    bool
	CloudBreakoutDown  bool
	Description        string
}

type IchimokuResult struct {
	Data     IchimokuData
	Analysis IchimokuAnalysis
}

func CalculateIchimoku(klines []AsterDexKline) IchimokuResult {
	if len(klines) < 52 {
		return IchimokuResult{
			Analysis: IchimokuAnalysis{
				Signal:      IchimokuSignalNeutral,
				Description: "Not enough data for Ichimoku calculation (need at least 52 candles)",
			},
		}
	}

	highs := make([]float64, len(klines))
	lows := make([]float64, len(klines))
	closes := make([]float64, len(klines))

	for i, k := range klines {
		highs[i], _ = strconv.ParseFloat(k.High, 64)
		lows[i], _ = strconv.ParseFloat(k.Low, 64)
		closes[i], _ = strconv.ParseFloat(k.Close, 64)
	}

	tenkan := calculateTenkan(highs, lows)
	kijun := calculateKijun(highs, lows)
	senkouA := calculateSenkouA(tenkan, kijun)
	senkouB := calculateSenkouB(highs, lows)
	chikou := calculateChikou(closes)

	data := IchimokuData{
		Tenkan:  convertToLines(klines, tenkan),
		Kijun:   convertToLines(klines, kijun),
		SenkouA: convertToLinesShifted(klines, senkouA, 26),
		SenkouB: convertToLinesShifted(klines, senkouB, 26),
		Chikou:  convertToLinesShifted(klines, chikou, -26),
		Price:   convertPriceToLines(klines),
	}

	analysis := analyzeIchimoku(closes, tenkan, kijun, senkouA, senkouB)

	return IchimokuResult{
		Data:     data,
		Analysis: analysis,
	}
}

func calculateTenkan(highs, lows []float64) []float64 {
	period := 9
	result := make([]float64, len(highs))

	for i := 0; i < len(highs); i++ {
		if i < period-1 {
			result[i] = 0
			continue
		}
		maxHigh := maxSlice(highs[i-period+1 : i+1])
		minLow := minSlice(lows[i-period+1 : i+1])
		result[i] = (maxHigh + minLow) / 2
	}
	return result
}

func calculateKijun(highs, lows []float64) []float64 {
	period := 26
	result := make([]float64, len(highs))

	for i := 0; i < len(highs); i++ {
		if i < period-1 {
			result[i] = 0
			continue
		}
		maxHigh := maxSlice(highs[i-period+1 : i+1])
		minLow := minSlice(lows[i-period+1 : i+1])
		result[i] = (maxHigh + minLow) / 2
	}
	return result
}

func calculateSenkouA(tenkan, kijun []float64) []float64 {
	result := make([]float64, len(tenkan))
	for i := 0; i < len(tenkan); i++ {
		if tenkan[i] == 0 || kijun[i] == 0 {
			result[i] = 0
			continue
		}
		result[i] = (tenkan[i] + kijun[i]) / 2
	}
	return result
}

func calculateSenkouB(highs, lows []float64) []float64 {
	period := 52
	result := make([]float64, len(highs))

	for i := 0; i < len(highs); i++ {
		if i < period-1 {
			result[i] = 0
			continue
		}
		maxHigh := maxSlice(highs[i-period+1 : i+1])
		minLow := minSlice(lows[i-period+1 : i+1])
		result[i] = (maxHigh + minLow) / 2
	}
	return result
}

func calculateChikou(closes []float64) []float64 {
	return closes
}

func analyzeIchimoku(closes, tenkan, kijun, senkouA, senkouB []float64) IchimokuAnalysis {
	n := len(closes)
	if n < 52 {
		return IchimokuAnalysis{
			Signal:      IchimokuSignalNeutral,
			Description: "Insufficient data",
		}
	}

	currentPrice := closes[n-1]
	prevPrice := closes[n-2]

	currentCloudTop := math.Max(senkouA[n-1], senkouB[n-1])
	currentCloudBottom := math.Min(senkouA[n-1], senkouB[n-1])

	prevCloudTop := math.Max(senkouA[n-2], senkouB[n-2])
	prevCloudBottom := math.Min(senkouA[n-2], senkouB[n-2])

	analysis := IchimokuAnalysis{}

	analysis.PriceAboveCloud = currentPrice > currentCloudTop
	analysis.PriceBelowCloud = currentPrice < currentCloudBottom
	analysis.PriceInCloud = !analysis.PriceAboveCloud && !analysis.PriceBelowCloud

	analysis.BullishCloud = senkouA[n-1] > senkouB[n-1]

	analysis.TenkanAboveKijun = tenkan[n-1] > kijun[n-1]

	analysis.TwoCloseAboveCloud = currentPrice > currentCloudTop && prevPrice > prevCloudTop
	analysis.TwoCloseBelowCloud = currentPrice < currentCloudBottom && prevPrice < prevCloudBottom

	prevPriceInCloud := prevPrice >= prevCloudBottom && prevPrice <= prevCloudTop
	analysis.CloudBreakoutUp = prevPriceInCloud && currentPrice > currentCloudTop
	analysis.CloudBreakoutDown = prevPriceInCloud && currentPrice < currentCloudBottom

	if analysis.CloudBreakoutUp && analysis.BullishCloud {
		analysis.Signal = IchimokuSignalStrongLong
		analysis.Description = "Strong LONG: Price broke out above cloud with bullish cloud color"
	} else if analysis.TwoCloseAboveCloud && analysis.TenkanAboveKijun {
		analysis.Signal = IchimokuSignalStrongLong
		analysis.Description = "Strong LONG: Two closes above cloud + Tenkan above Kijun"
	} else if analysis.PriceAboveCloud && analysis.TenkanAboveKijun {
		analysis.Signal = IchimokuSignalLong
		analysis.Description = "LONG: Price above cloud and Tenkan above Kijun"
	} else if analysis.PriceAboveCloud {
		analysis.Signal = IchimokuSignalLong
		analysis.Description = "LONG: Price above cloud"
	} else if analysis.CloudBreakoutDown && !analysis.BullishCloud {
		analysis.Signal = IchimokuSignalStrongShort
		analysis.Description = "Strong SHORT: Price broke down below cloud with bearish cloud color"
	} else if analysis.TwoCloseBelowCloud && !analysis.TenkanAboveKijun {
		analysis.Signal = IchimokuSignalStrongShort
		analysis.Description = "Strong SHORT: Two closes below cloud + Tenkan below Kijun"
	} else if analysis.PriceBelowCloud && !analysis.TenkanAboveKijun {
		analysis.Signal = IchimokuSignalShort
		analysis.Description = "SHORT: Price below cloud and Tenkan below Kijun"
	} else if analysis.PriceBelowCloud {
		analysis.Signal = IchimokuSignalShort
		analysis.Description = "SHORT: Price below cloud"
	} else if analysis.PriceInCloud {
		analysis.Signal = IchimokuSignalUncertain
		analysis.Description = "UNCERTAIN: Price is inside cloud - no clear signal"
	} else {
		analysis.Signal = IchimokuSignalNeutral
		analysis.Description = "NEUTRAL: No clear signal"
	}

	return analysis
}

func convertToLines(klines []AsterDexKline, values []float64) []IchimokuLine {
	lines := make([]IchimokuLine, len(klines))
	for i := range klines {
		lines[i] = IchimokuLine{
			Timestamp: klines[i].OpenTime,
			Value:     values[i],
		}
	}
	return lines
}

func convertToLinesShifted(klines []AsterDexKline, values []float64, shift int) []IchimokuLine {
	lines := make([]IchimokuLine, len(klines))
	for i := range klines {
		shiftedIndex := i + shift
		if shiftedIndex >= 0 && shiftedIndex < len(values) {
			lines[i] = IchimokuLine{
				Timestamp: klines[i].OpenTime,
				Value:     values[shiftedIndex],
			}
		}
	}
	return lines
}

func convertPriceToLines(klines []AsterDexKline) []IchimokuLine {
	lines := make([]IchimokuLine, len(klines))
	for i, k := range klines {
		close, _ := strconv.ParseFloat(k.Close, 64)
		lines[i] = IchimokuLine{
			Timestamp: k.OpenTime,
			Value:     close,
		}
	}
	return lines
}

func maxSlice(values []float64) float64 {
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func minSlice(values []float64) float64 {
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func GetTradingAction(analysis IchimokuAnalysis, currentPosition PositionSide) TradingAction {
	if currentPosition == PositionSideLong {
		if analysis.Signal == IchimokuSignalStrongShort || analysis.Signal == IchimokuSignalShort {
			return TradingActionCloseLong
		}
		return TradingActionHold
	}

	if currentPosition == PositionSideShort {
		if analysis.Signal == IchimokuSignalStrongLong || analysis.Signal == IchimokuSignalLong {
			return TradingActionCloseShort
		}
		return TradingActionHold
	}

	if analysis.Signal == IchimokuSignalStrongLong || analysis.Signal == IchimokuSignalLong {
		return TradingActionOpenLong
	}

	if analysis.Signal == IchimokuSignalStrongShort || analysis.Signal == IchimokuSignalShort {
		return TradingActionOpenShort
	}

	return TradingActionHold
}
