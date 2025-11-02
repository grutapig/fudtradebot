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

func ShouldClosePosition(positionSide PositionSide, ichimokuResult IchimokuResult) bool {
	n := len(ichimokuResult.Data.Price)
	if n < 2 {
		return false
	}

	currentPrice := ichimokuResult.Data.Price[n-1].Value
	currentKijun := ichimokuResult.Data.Kijun[n-1].Value
	currentTenkan := ichimokuResult.Data.Tenkan[n-1].Value

	currentCloudTop := math.Max(
		ichimokuResult.Data.SenkouA[n-1].Value,
		ichimokuResult.Data.SenkouB[n-1].Value,
	)
	currentCloudBottom := math.Min(
		ichimokuResult.Data.SenkouA[n-1].Value,
		ichimokuResult.Data.SenkouB[n-1].Value,
	)

	if positionSide == PositionSideLong {
		return shouldCloseLongPosition(currentPrice, currentKijun, currentTenkan, currentCloudTop, currentCloudBottom)
	} else if positionSide == PositionSideShort {
		return shouldCloseShortPosition(currentPrice, currentKijun, currentTenkan, currentCloudTop, currentCloudBottom)
	}

	return false
}

func shouldCloseLongPosition(price, kijun, tenkan, cloudTop, cloudBottom float64) bool {
	if price < kijun {
		return true
	}

	if price <= cloudTop && price >= cloudBottom {
		return true
	}

	if tenkan < kijun {
		return true
	}

	return false
}

func shouldCloseShortPosition(price, kijun, tenkan, cloudTop, cloudBottom float64) bool {
	if price > kijun {
		return true
	}

	if price <= cloudTop && price >= cloudBottom {
		return true
	}

	if tenkan > kijun {
		return true
	}

	return false
}

type ClosePositionReason struct {
	ReasonToClose    []string
	ReasonNotToClose []string
	ShouldClose      bool
	FinalExplanation string
}

func ShouldClosePositionDetailed(positionSide PositionSide, ichimokuResult IchimokuResult) ClosePositionReason {
	n := len(ichimokuResult.Data.Price)
	if n < 2 {
		return ClosePositionReason{
			ReasonNotToClose: []string{"Insufficient data for analysis"},
			ShouldClose:      false,
			FinalExplanation: "No data to make a decision",
		}
	}

	currentPrice := ichimokuResult.Data.Price[n-1].Value
	currentKijun := ichimokuResult.Data.Kijun[n-1].Value
	currentTenkan := ichimokuResult.Data.Tenkan[n-1].Value

	currentCloudTop := math.Max(
		ichimokuResult.Data.SenkouA[n-1].Value,
		ichimokuResult.Data.SenkouB[n-1].Value,
	)
	currentCloudBottom := math.Min(
		ichimokuResult.Data.SenkouA[n-1].Value,
		ichimokuResult.Data.SenkouB[n-1].Value,
	)

	if positionSide == PositionSideLong {
		return shouldCloseLongPositionDetailed(currentPrice, currentKijun, currentTenkan, currentCloudTop, currentCloudBottom)
	} else if positionSide == PositionSideShort {
		return shouldCloseShortPositionDetailed(currentPrice, currentKijun, currentTenkan, currentCloudTop, currentCloudBottom)
	}

	return ClosePositionReason{
		ReasonNotToClose: []string{"Unknown position type"},
		ShouldClose:      false,
		FinalExplanation: "Unable to determine position type",
	}
}

func shouldCloseLongPositionDetailed(price, kijun, tenkan, cloudTop, cloudBottom float64) ClosePositionReason {
	reasonsToClose := []string{}
	reasonsNotToClose := []string{}

	if price < kijun {
		reasonsToClose = append(reasonsToClose, "Price dropped below Kijun line - strong bearish signal")
	} else {
		reasonsNotToClose = append(reasonsNotToClose, "Price above Kijun - trend still maintained")
	}

	if price <= cloudTop && price >= cloudBottom {
		reasonsToClose = append(reasonsToClose, "Price entered the cloud - zone of uncertainty, weakening trend")
	} else if price > cloudTop {
		reasonsNotToClose = append(reasonsNotToClose, "Price above cloud - bullish trend continues")
	} else {
		reasonsToClose = append(reasonsToClose, "Price below cloud - trend reversed to bearish")
	}

	if tenkan < kijun {
		reasonsToClose = append(reasonsToClose, "Tenkan crossed below Kijun - bearish short-term trend signal")
	} else {
		reasonsNotToClose = append(reasonsNotToClose, "Tenkan above Kijun - short-term trend is bullish")
	}

	shouldClose := price < kijun || (price <= cloudTop && price >= cloudBottom) || tenkan < kijun

	finalExplanation := ""
	if shouldClose {
		finalExplanation = "Recommend closing long position: bearish signals detected indicating weakening or reversal of uptrend"
	} else {
		finalExplanation = "Long position can be held: bullish trend continues, key indicators show strength"
	}

	return ClosePositionReason{
		ReasonToClose:    reasonsToClose,
		ReasonNotToClose: reasonsNotToClose,
		ShouldClose:      shouldClose,
		FinalExplanation: finalExplanation,
	}
}

func shouldCloseShortPositionDetailed(price, kijun, tenkan, cloudTop, cloudBottom float64) ClosePositionReason {
	reasonsToClose := []string{}
	reasonsNotToClose := []string{}

	if price > kijun {
		reasonsToClose = append(reasonsToClose, "Price rose above Kijun line - strong bullish signal")
	} else {
		reasonsNotToClose = append(reasonsNotToClose, "Price below Kijun - downtrend continues")
	}

	if price <= cloudTop && price >= cloudBottom {
		reasonsToClose = append(reasonsToClose, "Price entered the cloud - zone of uncertainty, weakening trend")
	} else if price < cloudBottom {
		reasonsNotToClose = append(reasonsNotToClose, "Price below cloud - bearish trend continues")
	} else {
		reasonsToClose = append(reasonsToClose, "Price above cloud - trend reversed to bullish")
	}

	if tenkan > kijun {
		reasonsToClose = append(reasonsToClose, "Tenkan crossed above Kijun - bullish short-term trend signal")
	} else {
		reasonsNotToClose = append(reasonsNotToClose, "Tenkan below Kijun - short-term trend is bearish")
	}

	shouldClose := price > kijun || (price <= cloudTop && price >= cloudBottom) || tenkan > kijun

	finalExplanation := ""
	if shouldClose {
		finalExplanation = "Recommend closing short position: bullish signals detected indicating weakening or reversal of downtrend"
	} else {
		finalExplanation = "Short position can be held: bearish trend continues, key indicators show strength"
	}

	return ClosePositionReason{
		ReasonToClose:    reasonsToClose,
		ReasonNotToClose: reasonsNotToClose,
		ShouldClose:      shouldClose,
		FinalExplanation: finalExplanation,
	}
}
