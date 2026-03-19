package trading212

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DividendCashAction controls what happens to dividend cash inside a Pie.
type DividendCashAction string

const (
	// DividendCashActionReinvest automatically reinvests dividends.
	DividendCashActionReinvest DividendCashAction = "REINVEST"

	// DividendCashActionToAccount sends dividends to the main account balance.
	DividendCashActionToAccount DividendCashAction = "TO_ACCOUNT_CASH"
)

// PieInstrument is an instrument (slice) within a Pie.
type PieInstrument struct {
	// Ticker is the instrument ticker.
	Ticker string `json:"ticker"`

	// Result holds P&L information for this slice.
	Result PieInstrumentResult `json:"result"`

	// OwnedQuantity is the number of shares held in this slice.
	OwnedQuantity float64 `json:"ownedQuantity"`

	// CurrentShare is the actual allocation percentage (0–100).
	CurrentShare float64 `json:"currentShare"`

	// ExpectedShare is the target allocation percentage (0–100) set by the
	// user.
	ExpectedShare float64 `json:"expectedShare"`
}

// PieInstrumentResult holds the P&L breakdown for a Pie slice.
type PieInstrumentResult struct {
	// InvestedValue is the total amount invested in this slice.
	InvestedValue float64 `json:"investedValue"`

	// Value is the current market value of this slice.
	Value float64 `json:"value"`

	// PPL is the unrealised profit / loss for this slice.
	PPL float64 `json:"result"`
}

// PieSummary is a lightweight Pie as returned in list responses.
type PieSummary struct {
	// ID is the unique Pie identifier.
	ID int64 `json:"id"`

	// Cash is the uninvested cash balance within the Pie.
	Cash float64 `json:"cash"`

	// DividendCashAction controls dividend handling.
	DividendCashAction DividendCashAction `json:"dividendCashAction"`

	// EndDate is the optional target date for the Pie (nil if not set).
	EndDate *time.Time `json:"endDate,omitempty"`

	// GoalAmount is the optional investment goal amount (nil if not set).
	GoalAmount *float64 `json:"goal,omitempty"`

	// Icon is the emoji or icon identifier chosen for the Pie.
	Icon string `json:"icon,omitempty"`

	// InstrumentShares maps ticker → target share percentage (0–100).
	InstrumentShares map[string]float64 `json:"instrumentShares"`

	// Name is the user-defined Pie name.
	Name string `json:"name"`
}

// Pie is a fully detailed Pie including instrument-level data.
type Pie struct {
	// Settings holds the Pie configuration.
	Settings PieSummary `json:"settings"`

	// Instruments lists the individual slices with performance data.
	Instruments []PieInstrument `json:"instruments"`
}

// CreatePieRequest is the request body for [Client.CreatePie].
type CreatePieRequest struct {
	// DividendCashAction controls dividend handling. Required.
	DividendCashAction DividendCashAction `json:"dividendCashAction"`

	// EndDate is an optional target date for the Pie.
	EndDate *time.Time `json:"endDate,omitempty"`

	// GoalAmount is an optional investment goal amount.
	GoalAmount *float64 `json:"goal,omitempty"`

	// Icon is an optional emoji or icon identifier.
	Icon string `json:"icon,omitempty"`

	// InstrumentShares maps ticker → target share percentage (0–100). Required.
	InstrumentShares map[string]float64 `json:"instrumentShares"`

	// Name is the Pie name. Required.
	Name string `json:"name"`
}

// UpdatePieRequest is the request body for [Client.UpdatePie].
type UpdatePieRequest = CreatePieRequest

// ---------------------------------------------------------------------------
// Pie methods
// ---------------------------------------------------------------------------

// GetPies returns a summary of all Pies in the authenticated account.
//
// Rate limit: 1 request / 5 s.
func (c *Client) GetPies(ctx context.Context) ([]PieSummary, error) {
	var out []PieSummary
	_, err := c.do(ctx, http.MethodGet, "/equity/pies", nil, &out)
	return out, err
}

// CreatePie creates a new Pie.
//
// Rate limit: 1 request / 5 s.
func (c *Client) CreatePie(ctx context.Context, req CreatePieRequest) (Pie, error) {
	var out Pie
	_, err := c.do(ctx, http.MethodPost, "/equity/pies", req, &out)
	return out, err
}

// GetPie returns the full details of a Pie by its ID.
//
// Rate limit: 1 request / 5 s.
func (c *Client) GetPie(ctx context.Context, id int64) (Pie, error) {
	var out Pie
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/equity/pies/%d", id), nil, &out)
	return out, err
}

// UpdatePie replaces the settings of an existing Pie.
//
// Rate limit: 1 request / 5 s.
func (c *Client) UpdatePie(ctx context.Context, id int64, req UpdatePieRequest) (Pie, error) {
	var out Pie
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/equity/pies/%d", id), req, &out)
	return out, err
}

// DeletePie deletes a Pie by its ID.
//
// Rate limit: 1 request / 30 s.
func (c *Client) DeletePie(ctx context.Context, id int64) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/equity/pies/%d", id), nil, nil)
	return err
}

// DuplicatePie creates an identical copy of an existing Pie.
//
// Rate limit: 1 request / 30 s.
func (c *Client) DuplicatePie(ctx context.Context, id int64) (Pie, error) {
	var out Pie
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/equity/pies/%d/duplicate", id), nil, &out)
	return out, err
}
