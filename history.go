package trading212

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HistoryOrder is an order from the account's order history.
type HistoryOrder struct {
	// ID is the unique order identifier.
	ID string `json:"id"`

	// Ticker is the instrument ticker.
	Ticker string `json:"ticker"`

	// Type is the order type.
	Type OrderType `json:"type"`

	// Status is the order status at settlement.
	Status OrderStatus `json:"status"`

	// Side is BUY or SELL.
	Side OrderSide `json:"side"`

	// Quantity is the total ordered quantity.
	Quantity float64 `json:"quantity"`

	// FilledQuantity is the quantity that was executed.
	FilledQuantity float64 `json:"filledQuantity"`

	// LimitPrice is the limit price (LIMIT and STOP_LIMIT orders).
	LimitPrice *float64 `json:"limitPrice,omitempty"`

	// StopPrice is the stop trigger price (STOP and STOP_LIMIT orders).
	StopPrice *float64 `json:"stopPrice,omitempty"`

	// FillPrice is the average price at which the order was executed.
	FillPrice *float64 `json:"fillPrice,omitempty"`

	// DateCreated is when the order was placed.
	DateCreated time.Time `json:"dateCreated"`

	// DateModified is when the order was last modified.
	DateModified *time.Time `json:"dateModified,omitempty"`

	// Taxes lists any taxes applied to this order.
	Taxes []Tax `json:"taxes"`
}

// Tax is a tax line item on an order.
type Tax struct {
	// Name is the tax name (e.g. "FinancialTransactionTax").
	Name string `json:"name"`

	// Quantity is the tax amount in account currency.
	Quantity float64 `json:"quantity"`

	// TimeCredited is when the tax was applied.
	TimeCredited time.Time `json:"timeCredited"`
}

// Dividend is a dividend payment recorded in the account history.
type Dividend struct {
	// Ticker is the instrument ticker.
	Ticker string `json:"ticker"`

	// Quantity is the number of shares that generated the dividend.
	Quantity float64 `json:"quantity"`

	// Amount is the gross dividend amount in account currency.
	Amount float64 `json:"amount"`

	// AmountInEuro is the gross dividend amount in EUR.
	AmountInEuro float64 `json:"amountInEuro"`

	// GrossAmountPerShare is the per-share gross dividend.
	GrossAmountPerShare float64 `json:"grossAmountPerShare"`

	// PaidOn is the date the dividend was paid.
	PaidOn time.Time `json:"paidOn"`

	// Type is the dividend type (e.g. "ORDINARY").
	Type string `json:"type"`
}

// Transaction is a cash transaction in the account history.
type Transaction struct {
	// Reference is a unique identifier for the transaction.
	Reference string `json:"reference"`

	// Amount is the transaction amount in account currency. Negative for
	// withdrawals / debits.
	Amount float64 `json:"amount"`

	// Type describes the transaction category (e.g. "DEPOSIT", "WITHDRAWAL").
	Type string `json:"type"`

	// DateTime is when the transaction was processed.
	DateTime time.Time `json:"dateTime"`
}

// ExportStatus enumerates the possible states of a history export.
type ExportStatus string

const (
	ExportStatusProcessing ExportStatus = "Processing"
	ExportStatusReady      ExportStatus = "Finished"
	ExportStatusFailed     ExportStatus = "Failed"
)

// Export represents a CSV export report.
type Export struct {
	// ReportID is the unique report identifier returned when the export was
	// requested.
	ReportID int64 `json:"reportId"`

	// Status is the current export status.
	Status ExportStatus `json:"status"`

	// DownloadLink is the URL for the CSV download (populated when Status is
	// [ExportStatusReady]).
	DownloadLink string `json:"downloadLink,omitempty"`

	// DataFrom is the start of the exported date range.
	DataFrom time.Time `json:"dataFrom"`

	// DataTo is the end of the exported date range.
	DataTo time.Time `json:"dataTo"`
}

// ExportRequest is the request body for [Client.RequestExport].
type ExportRequest struct {
	// DataIncluded specifies which data types to include in the export.
	DataIncluded ExportDataIncluded `json:"dataIncluded"`

	// TimeFrom is the start of the date range to export.
	TimeFrom time.Time `json:"timeFrom"`

	// TimeTo is the end of the date range to export.
	TimeTo time.Time `json:"timeTo"`
}

// ExportDataIncluded selects which transaction types to include in an export.
type ExportDataIncluded struct {
	// IncludeDividends includes dividend payments.
	IncludeDividends bool `json:"includeDividends"`

	// IncludeInterest includes interest payments.
	IncludeInterest bool `json:"includeInterest"`

	// IncludeOrders includes order history.
	IncludeOrders bool `json:"includeOrders"`

	// IncludeTransactions includes cash transactions.
	IncludeTransactions bool `json:"includeTransactions"`
}

// ---------------------------------------------------------------------------
// Params types
// ---------------------------------------------------------------------------

// HistoryOrdersParams configures a paginated history orders query.
type HistoryOrdersParams struct {
	// Limit sets the number of items per page (max 50, 0 = API default of 20).
	Limit int

	// Ticker filters results to a single instrument.
	Ticker string
}

// HistoryDividendsParams configures a paginated dividend history query.
type HistoryDividendsParams struct {
	// Limit sets the number of items per page (max 50, 0 = API default of 20).
	Limit int

	// Ticker filters results to a single instrument.
	Ticker string
}

// HistoryTransactionsParams configures a paginated transaction history query.
type HistoryTransactionsParams struct {
	// Limit sets the number of items per page (max 50, 0 = API default of 20).
	Limit int

	// Cursor is the pagination cursor; leave empty to start from the beginning.
	Cursor string
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildHistoryQuery(limit int, ticker, cursor string) string {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if ticker != "" {
		q.Set("ticker", ticker)
	}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

// ---------------------------------------------------------------------------
// History methods
// ---------------------------------------------------------------------------

// HistoryOrders returns a cursor over the account's order history.
//
// Rate limit: 6 requests / 60 s.
func (c *Client) HistoryOrders(params HistoryOrdersParams) *Cursor[HistoryOrder] {
	return newCursor(func(ctx context.Context, cursor string) (Page[HistoryOrder], error) {
		path := "/equity/history/orders" + buildHistoryQuery(params.Limit, params.Ticker, cursor)
		var page Page[HistoryOrder]
		_, err := c.do(ctx, http.MethodGet, path, nil, &page)
		return page, err
	})
}

// HistoryDividends returns a cursor over the account's dividend history.
//
// Rate limit: 6 requests / 60 s.
func (c *Client) HistoryDividends(params HistoryDividendsParams) *Cursor[Dividend] {
	return newCursor(func(ctx context.Context, cursor string) (Page[Dividend], error) {
		path := "/equity/history/dividends" + buildHistoryQuery(params.Limit, params.Ticker, cursor)
		var page Page[Dividend]
		_, err := c.do(ctx, http.MethodGet, path, nil, &page)
		return page, err
	})
}

// HistoryTransactions returns a cursor over the account's transaction history.
//
// Rate limit: 6 requests / 60 s.
func (c *Client) HistoryTransactions(params HistoryTransactionsParams) *Cursor[Transaction] {
	return newCursor(func(ctx context.Context, cursor string) (Page[Transaction], error) {
		path := "/equity/history/transactions" + buildHistoryQuery(params.Limit, "", cursor)
		var page Page[Transaction]
		_, err := c.do(ctx, http.MethodGet, path, nil, &page)
		return page, err
	})
}

// RequestExport triggers generation of a CSV export report.
// Use [Client.GetExports] to poll the returned report ID for completion.
//
// Rate limit: 1 request / 30 s.
func (c *Client) RequestExport(ctx context.Context, req ExportRequest) (Export, error) {
	var out Export
	_, err := c.do(ctx, http.MethodPost, "/equity/history/exports", req, &out)
	return out, err
}

// GetExports returns the status of all requested CSV exports.
//
// Rate limit: 1 request / 60 s.
func (c *Client) GetExports(ctx context.Context) ([]Export, error) {
	var out []Export
	_, err := c.do(ctx, http.MethodGet, "/equity/history/exports", nil, &out)
	return out, err
}

// GetExport returns the status of a specific export by its report ID.
func (c *Client) GetExport(ctx context.Context, reportID int64) (Export, error) {
	exports, err := c.GetExports(ctx)
	if err != nil {
		return Export{}, err
	}
	for _, e := range exports {
		if e.ReportID == reportID {
			return e, nil
		}
	}
	return Export{}, fmt.Errorf("trading212: export %d not found", reportID)
}
