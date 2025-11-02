package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func GenerateCandlestickSVG(klines []AsterDexKline, width, height int) string {
	if len(klines) == 0 {
		return ""
	}

	padding := 50
	chartWidth := width - 2*padding
	chartHeight := height - 2*padding

	minPrice, maxPrice := findPriceRange(klines)
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	candleWidth := float64(chartWidth) / float64(len(klines))
	bodyWidth := candleWidth * 0.7

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#1a1a1a"/>`, width, height))

	for i, kline := range klines {
		open, _ := strconv.ParseFloat(kline.Open, 64)
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)
		close, _ := strconv.ParseFloat(kline.Close, 64)

		x := float64(padding) + float64(i)*candleWidth + candleWidth/2

		highY := float64(padding) + float64(chartHeight)*(1-(high-minPrice)/priceRange)
		lowY := float64(padding) + float64(chartHeight)*(1-(low-minPrice)/priceRange)
		openY := float64(padding) + float64(chartHeight)*(1-(open-minPrice)/priceRange)
		closeY := float64(padding) + float64(chartHeight)*(1-(close-minPrice)/priceRange)

		isGreen := close >= open
		color := "#ef4444"
		if isGreen {
			color = "#22c55e"
		}

		svg.WriteString(fmt.Sprintf(`<line x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="%s" stroke-width="1"/>`,
			x, highY, x, lowY, color))

		bodyTop := math.Min(openY, closeY)
		bodyHeight := math.Abs(closeY - openY)
		if bodyHeight < 1 {
			bodyHeight = 1
		}

		svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s"/>`,
			x-bodyWidth/2, bodyTop, bodyWidth, bodyHeight, color))
	}

	svg.WriteString(`</svg>`)
	return svg.String()
}

func findPriceRange(klines []AsterDexKline) (float64, float64) {
	minPrice := math.MaxFloat64
	maxPrice := -math.MaxFloat64

	for _, kline := range klines {
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)

		if high > maxPrice {
			maxPrice = high
		}
		if low < minPrice {
			minPrice = low
		}
	}

	return minPrice, maxPrice
}
