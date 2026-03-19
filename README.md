# trading-212-go

A Go client library for the [Trading 212 REST API](https://t212public-api-docs.redoc.ly/).

[![Go Reference](https://pkg.go.dev/badge/github.com/joaoamatos/trading-212-go.svg)](https://pkg.go.dev/github.com/joaoamatos/trading-212-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/joaoamatos/trading-212-go)](https://goreportcard.com/report/github.com/joaoamatos/trading-212-go)

## Requirements

- Go 1.21+
- A Trading 212 **Invest** or **Stocks ISA** account
- An API key from **Settings → API (Beta)**

> **Note:** The Trading 212 API is currently in beta and available for Invest and Stocks ISA accounts only.

## Installation

```sh
go get github.com/joaoamatos/trading-212-go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"

    trading212 "github.com/joaoamatos/trading-212-go"
)

func main() {
    client := trading212.New("your-api-key")

    ctx := context.Background()

    summary, err := client.GetAccountSummary(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Account %d — free cash: %.2f %s\n",
        summary.ID, summary.Cash.Free, summary.CurrencyCode)
}
```

Use `WithDemo()` to target the paper-trading environment:

```go
client := trading212.New("your-demo-api-key", trading212.WithDemo())
```

## Usage

### Account

```go
summary, err := client.GetAccountSummary(ctx)
```

### Orders

```go
// List pending orders
orders, err := client.GetOrders(ctx)

// Place a market order (buy 5 shares of Apple)
order, err := client.PlaceMarketOrder(ctx, trading212.MarketOrderRequest{
    Ticker:   "AAPL_US_EQ",
    Quantity: 5,
})

// Place a limit order
order, err := client.PlaceLimitOrder(ctx, trading212.LimitOrderRequest{
    Ticker:       "AAPL_US_EQ",
    Quantity:     5,
    LimitPrice:   180.00,
    TimeValidity: trading212.TimeValidityDay,
})

// Cancel a pending order
err = client.CancelOrder(ctx, order.ID)
```

### Positions

```go
positions, err := client.GetPositions(ctx)
```

### Instruments & Exchanges

```go
instruments, err := client.GetInstruments(ctx)
exchanges, err := client.GetExchanges(ctx)
```

### History (paginated)

History endpoints return a `*Cursor[T]` that fetches pages lazily. Use it like
a scanner:

```go
cur := client.HistoryOrders(trading212.HistoryOrdersParams{
    Limit:  50,
    Ticker: "AAPL_US_EQ", // optional filter
})

for cur.Next(ctx) {
    order := cur.Item()
    fmt.Println(order.ID, order.Status)
}
if err := cur.Err(); err != nil {
    log.Fatal(err)
}
```

The same pattern applies to dividends and transactions:

```go
cur := client.HistoryDividends(trading212.HistoryDividendsParams{Limit: 50})
cur := client.HistoryTransactions(trading212.HistoryTransactionsParams{Limit: 50})
```

### Exports

```go
// Request a CSV export
export, err := client.RequestExport(ctx, trading212.ExportRequest{
    TimeFrom: time.Now().AddDate(0, -1, 0),
    TimeTo:   time.Now(),
    DataIncluded: trading212.ExportDataIncluded{
        IncludeOrders:       true,
        IncludeDividends:    true,
        IncludeTransactions: true,
    },
})

// Poll until ready
export, err = client.GetExport(ctx, export.ReportID)
if export.Status == trading212.ExportStatusReady {
    fmt.Println("Download:", export.DownloadLink)
}
```

### Pies

```go
// List all pies
pies, err := client.GetPies(ctx)

// Create a pie
pie, err := client.CreatePie(ctx, trading212.CreatePieRequest{
    Name:               "Tech stocks",
    DividendCashAction: trading212.DividendCashActionReinvest,
    InstrumentShares: map[string]float64{
        "AAPL_US_EQ": 50,
        "MSFT_US_EQ": 50,
    },
})

// Delete a pie
err = client.DeletePie(ctx, pie.Settings.ID)
```

## Error handling

All methods return a plain `error`. HTTP-level errors (status ≥ 400) are
returned as `*trading212.Error`, which carries the status code and the message
from the API response body:

```go
_, err := client.GetAccountSummary(ctx)

var apiErr *trading212.Error
if errors.As(err, &apiErr) {
    switch apiErr.StatusCode {
    case 401:
        log.Fatal("invalid API key")
    case 429:
        fmt.Printf("rate limited — retry after %v\n", apiErr.RateLimit.RetryAfter)
    default:
        log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Message)
    }
}
```

Network-level failures (DNS, connection refused, timeout) are returned as
standard `net/http` errors and are never wrapped in `*trading212.Error`.

## Configuration

| Option | Description |
|---|---|
| `WithDemo()` | Use the demo (paper trading) environment |
| `WithRateLimiting()` | Enable automatic client-side rate limiting (see below) |
| `WithHTTPClient(hc)` | Replace the default HTTP client (add logging, retries, etc.) |
| `WithTimeout(d)` | Set a custom request timeout (default: 30s) |
| `WithBaseURL(u)` | Override the base URL (useful for testing) |

## Rate limits

Trading 212 enforces per-account rate limits. The library handles this in two
complementary ways.

### Automatic rate limiting (recommended)

Pass `WithRateLimiting()` when creating the client. The client will
transparently wait before sending a request that would exceed the documented
limit, using a per-endpoint token bucket. Context cancellation is fully
respected — if `ctx` is cancelled while a call is waiting for a slot, the
call returns `ctx.Err()` immediately.

```go
client := trading212.New("your-api-key", trading212.WithRateLimiting())

// These three calls are spaced automatically — no 429s, no manual sleeps.
for i := 0; i < 10; i++ {
    positions, err := client.GetPositions(ctx)
    _ = positions
}
```

Limits are tracked per `Client` instance. If you create multiple clients with
the same API key, each tracks its own budget independently — prefer a single
shared client.

### Manual rate limit handling

Without `WithRateLimiting`, rate-limit errors are returned as `*Error` with
`StatusCode == 429`. The `RateLimit` field carries the suggested wait duration:

```go
var apiErr *trading212.Error
if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
    time.Sleep(apiErr.RateLimit.RetryAfter)
}
```

You can also use `trading212.RateLimitFromResponse` with a custom HTTP
transport to inspect rate-limit headers on every response.

### Documented limits

| Endpoint | Limit |
|---|---|
| Account summary | 1 req / 5 s |
| Place market order | 50 req / 60 s |
| Place limit / stop order | 1 req / 2 s |
| Cancel order | 50 req / 60 s |
| Get positions | 1 req / 1 s |
| History (orders / dividends / transactions) | 6 req / 60 s |
| Instruments list | 1 req / 50 s |
| Exchanges | 1 req / 30 s |
| Request export | 1 req / 30 s |
| Get exports | 1 req / 60 s |

## License

[MIT](LICENSE)
