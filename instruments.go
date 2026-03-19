package trading212

import (
	"context"
	"net/http"
)

// InstrumentType enumerates the supported security types.
type InstrumentType string

const (
	InstrumentTypeStock InstrumentType = "STOCK"
	InstrumentTypeETF   InstrumentType = "ETF"
)

// WorkingSchedule describes when an exchange is open for trading.
type WorkingSchedule struct {
	// ID is the schedule identifier.
	ID int64 `json:"id"`

	// Phases lists the individual trading phases within the schedule.
	Phases []TradingPhase `json:"phases"`
}

// TradingPhase is a single phase within a working schedule.
type TradingPhase struct {
	// DateFrom is the date from which this phase applies (ISO 8601 date string).
	DateFrom string `json:"dateFrom,omitempty"`

	// DateTo is the date until which this phase applies (ISO 8601 date string).
	DateTo string `json:"dateTo,omitempty"`

	// StartTime is the phase start time (HH:MM:SS).
	StartTime string `json:"startTime"`

	// EndTime is the phase end time (HH:MM:SS).
	EndTime string `json:"endTime"`

	// Type identifies the phase type (e.g. "PRE_MARKET", "REGULAR", "AFTER_HOURS").
	Type string `json:"type"`
}

// Instrument describes a tradable security on Trading 212.
type Instrument struct {
	// Ticker is the unique Trading 212 ticker (e.g. "AAPL_US_EQ").
	Ticker string `json:"ticker"`

	// ShortName is the abbreviated security name.
	ShortName string `json:"shortName"`

	// Name is the full security name.
	Name string `json:"name"`

	// Type is the instrument type (STOCK, ETF, …).
	Type InstrumentType `json:"type"`

	// ISIN is the International Securities Identification Number.
	ISIN string `json:"isin"`

	// CurrencyCode is the ISO 4217 code for the instrument's trading currency.
	CurrencyCode string `json:"currencyCode"`

	// MinTradeQuantity is the minimum number of shares per order.
	MinTradeQuantity float64 `json:"minTradeQuantity"`

	// MaxOpenQuantity is the maximum shares that can be held simultaneously.
	MaxOpenQuantity *float64 `json:"maxOpenQuantity,omitempty"`

	// AddedOn is the date the instrument was added to the platform (ISO 8601).
	AddedOn string `json:"addedOn"`

	// Exchanges is the list of exchange codes where this instrument trades.
	Exchanges []string `json:"exchangesList"`

	// WorkingScheduleID is the ID of the applicable working schedule.
	WorkingScheduleID int64 `json:"workingScheduleId"`
}

// Exchange describes a stock exchange and its trading schedules.
type Exchange struct {
	// ID is the internal exchange identifier.
	ID int64 `json:"id"`

	// Name is the exchange name (e.g. "NASDAQ").
	Name string `json:"name"`

	// WorkingSchedules lists the schedules that govern trading hours.
	WorkingSchedules []WorkingSchedule `json:"workingSchedules"`
}

// GetInstruments returns all instruments available for trading.
//
// Rate limit: 1 request / 50 s.
func (c *Client) GetInstruments(ctx context.Context) ([]Instrument, error) {
	var out []Instrument
	_, err := c.do(ctx, http.MethodGet, "/equity/metadata/instruments", nil, &out)
	return out, err
}

// GetExchanges returns exchange information including trading schedules.
//
// Rate limit: 1 request / 30 s.
func (c *Client) GetExchanges(ctx context.Context) ([]Exchange, error) {
	var out []Exchange
	_, err := c.do(ctx, http.MethodGet, "/equity/metadata/exchanges", nil, &out)
	return out, err
}
