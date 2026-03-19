package trading212

import (
	"context"
	"net/http"
)

// Cash holds the cash breakdown from an account summary.
type Cash struct {
	// Blocked is the amount blocked (e.g. reserved for open orders).
	Blocked float64 `json:"blocked"`

	// Free is the cash available to trade.
	Free float64 `json:"free"`

	// Invested is the total amount currently invested.
	Invested float64 `json:"invested"`

	// PieCash is the cash held inside Pies.
	PieCash float64 `json:"pieCash"`

	// PPL is the profit/loss of the portfolio.
	PPL float64 `json:"ppl"`

	// Result is the total result (realised + unrealised P&L).
	Result float64 `json:"result"`

	// Total is the total account value (cash + investments).
	Total float64 `json:"total"`
}

// AccountSummary is the response from [Client.GetAccountSummary].
type AccountSummary struct {
	// ID is the unique account identifier.
	ID int64 `json:"id"`

	// CurrencyCode is the ISO 4217 currency code of the account.
	CurrencyCode string `json:"currencyCode"`

	// Cash holds the detailed cash breakdown.
	Cash Cash `json:"cash"`
}

// GetAccountSummary returns a cash breakdown and investment metrics for the
// authenticated account.
//
// Rate limit: 1 request / 5 s.
func (c *Client) GetAccountSummary(ctx context.Context) (AccountSummary, error) {
	var out AccountSummary
	_, err := c.do(ctx, http.MethodGet, "/equity/account/summary", nil, &out)
	return out, err
}
