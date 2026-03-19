package trading212

import (
	"context"
	"net/http"
	"time"
)

// Position represents an open equity position in the account.
type Position struct {
	// Ticker is the instrument ticker (e.g. "AAPL_US_EQ").
	Ticker string `json:"ticker"`

	// Quantity is the total number of shares held.
	Quantity float64 `json:"quantity"`

	// AveragePrice is the average price paid per share.
	AveragePrice float64 `json:"averagePrice"`

	// CurrentPrice is the latest market price per share.
	CurrentPrice float64 `json:"currentPrice"`

	// PPL is the unrealised profit / loss in account currency.
	PPL float64 `json:"ppl"`

	// FxPPL is the foreign-exchange component of unrealised P&L.
	FxPPL *float64 `json:"fxPpl,omitempty"`

	// InitialFillDate is when the position was first opened.
	InitialFillDate time.Time `json:"initialFillDate"`

	// QuantityAvailableForTrading is the portion not reserved for pending
	// orders.
	QuantityAvailableForTrading float64 `json:"maxSell"`

	// PieQuantity is the number of shares held inside Pies.
	PieQuantity float64 `json:"pieQuantity"`
}

// GetPositions returns all open positions for the authenticated account.
//
// Rate limit: 1 request / 1 s.
func (c *Client) GetPositions(ctx context.Context) ([]Position, error) {
	var out []Position
	_, err := c.do(ctx, http.MethodGet, "/equity/positions", nil, &out)
	return out, err
}
