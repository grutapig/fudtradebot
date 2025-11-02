package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func GenerateIchimokuSVG(klines []AsterDexKline, ichimoku IchimokuData, width, height int) string {
	if len(klines) == 0 {
		return ""
	}

	padding := 60
	chartWidth := width - 2*padding
	chartHeight := height - 2*padding

	minPrice, maxPrice := findPriceRangeWithIchimoku(klines, ichimoku)
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	candleWidth := float64(chartWidth) / float64(len(klines))
	bodyWidth := candleWidth * 0.6

	var svg strings.Builder
	svg.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#0a0a0a"/>`, width, height))

	drawCloud(&svg, ichimoku.SenkouA, ichimoku.SenkouB, minPrice, priceRange, padding, chartWidth, chartHeight, candleWidth)

	drawLine(&svg, ichimoku.Kijun, minPrice, priceRange, padding, chartWidth, chartHeight, candleWidth, "#9333ea", 2)
	drawLine(&svg, ichimoku.Tenkan, minPrice, priceRange, padding, chartWidth, chartHeight, candleWidth, "#06b6d4", 2)

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

		svg.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" opacity="0.9"/>`,
			x-bodyWidth/2, bodyTop, bodyWidth, bodyHeight, color))
	}

	svg.WriteString(`<text x="10" y="20" fill="#9333ea" font-size="12" font-family="monospace">Kijun (26)</text>`)
	svg.WriteString(`<text x="10" y="35" fill="#06b6d4" font-size="12" font-family="monospace">Tenkan (9)</text>`)
	svg.WriteString(`<text x="10" y="50" fill="#22c55e" font-size="12" font-family="monospace" opacity="0.5">Cloud</text>`)

	svg.WriteString(`</svg>`)
	return svg.String()
}

func drawCloud(svg *strings.Builder, senkouA, senkouB []IchimokuLine, minPrice, priceRange float64, padding, chartWidth, chartHeight int, candleWidth float64) {
	if len(senkouA) == 0 || len(senkouB) == 0 {
		return
	}

	var pathGreen, pathRed strings.Builder

	for i := 0; i < len(senkouA); i++ {
		if senkouA[i].Value == 0 || senkouB[i].Value == 0 {
			continue
		}

		x := float64(padding) + float64(i)*candleWidth + candleWidth/2
		yA := float64(padding) + float64(chartHeight)*(1-(senkouA[i].Value-minPrice)/priceRange)

		if i == 0 {
			pathGreen.WriteString(fmt.Sprintf("M %.2f %.2f", x, yA))
			pathRed.WriteString(fmt.Sprintf("M %.2f %.2f", x, yA))
		}

		isBullish := senkouA[i].Value > senkouB[i].Value

		if isBullish {
			if pathGreen.Len() == 0 || strings.Contains(pathGreen.String(), "Z") {
				pathGreen.Reset()
				pathGreen.WriteString(fmt.Sprintf("M %.2f %.2f", x, yA))
			}
			pathGreen.WriteString(fmt.Sprintf(" L %.2f %.2f", x, yA))
		} else {
			if pathRed.Len() == 0 || strings.Contains(pathRed.String(), "Z") {
				pathRed.Reset()
				pathRed.WriteString(fmt.Sprintf("M %.2f %.2f", x, yA))
			}
			pathRed.WriteString(fmt.Sprintf(" L %.2f %.2f", x, yA))
		}
	}

	for i := len(senkouB) - 1; i >= 0; i-- {
		if senkouA[i].Value == 0 || senkouB[i].Value == 0 {
			continue
		}

		x := float64(padding) + float64(i)*candleWidth + candleWidth/2
		yB := float64(padding) + float64(chartHeight)*(1-(senkouB[i].Value-minPrice)/priceRange)

		isBullish := senkouA[i].Value > senkouB[i].Value

		if isBullish {
			pathGreen.WriteString(fmt.Sprintf(" L %.2f %.2f", x, yB))
		} else {
			pathRed.WriteString(fmt.Sprintf(" L %.2f %.2f", x, yB))
		}
	}

	if pathGreen.Len() > 0 {
		pathGreen.WriteString(" Z")
		svg.WriteString(fmt.Sprintf(`<path d="%s" fill="#22c55e" opacity="0.2" stroke="none"/>`, pathGreen.String()))
	}

	if pathRed.Len() > 0 {
		pathRed.WriteString(" Z")
		svg.WriteString(fmt.Sprintf(`<path d="%s" fill="#ef4444" opacity="0.2" stroke="none"/>`, pathRed.String()))
	}
}

func drawLine(svg *strings.Builder, line []IchimokuLine, minPrice, priceRange float64, padding, chartWidth, chartHeight int, candleWidth float64, color string, strokeWidth int) {
	if len(line) == 0 {
		return
	}

	var path strings.Builder
	started := false

	for i, point := range line {
		if point.Value == 0 {
			continue
		}

		x := float64(padding) + float64(i)*candleWidth + candleWidth/2
		y := float64(padding) + float64(chartHeight)*(1-(point.Value-minPrice)/priceRange)

		if !started {
			path.WriteString(fmt.Sprintf("M %.2f %.2f", x, y))
			started = true
		} else {
			path.WriteString(fmt.Sprintf(" L %.2f %.2f", x, y))
		}
	}

	if path.Len() > 0 {
		svg.WriteString(fmt.Sprintf(`<path d="%s" fill="none" stroke="%s" stroke-width="%d" opacity="0.8"/>`, path.String(), color, strokeWidth))
	}
}

func findPriceRangeWithIchimoku(klines []AsterDexKline, ichimoku IchimokuData) (float64, float64) {
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

	for _, line := range ichimoku.SenkouA {
		if line.Value > 0 && line.Value > maxPrice {
			maxPrice = line.Value
		}
		if line.Value > 0 && line.Value < minPrice {
			minPrice = line.Value
		}
	}

	for _, line := range ichimoku.SenkouB {
		if line.Value > 0 && line.Value > maxPrice {
			maxPrice = line.Value
		}
		if line.Value > 0 && line.Value < minPrice {
			minPrice = line.Value
		}
	}

	return minPrice, maxPrice
}
