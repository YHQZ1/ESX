package channels

import "fmt"

func OrderBook(symbol string) string {
	return fmt.Sprintf("orderbook.%s", symbol)
}

func Trades(symbol string) string {
	return fmt.Sprintf("trades.%s", symbol)
}

func Ticker(symbol string) string {
	return fmt.Sprintf("ticker.%s", symbol)
}

func Candles(symbol, interval string) string {
	return fmt.Sprintf("candles.%s.%s", symbol, interval)
}
