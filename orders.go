package trading212

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// OrderType enumerates the supported order types.
type OrderType string

const (
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

// OrderSide indicates the direction of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderStatus describes the current lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "PENDING"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
)

// TimeValidity controls how long an order remains active.
type TimeValidity string

const (
	// TimeValidityDay expires at the end of the trading day.
	TimeValidityDay TimeValidity = "DAY"

	// TimeValidityGoodTillCancel remains active until explicitly cancelled.
	TimeValidityGoodTillCancel TimeValidity = "GOOD_TILL_CANCEL"
)

// Order represents an equity order (pending or historical).
type Order struct {
	// ID is the unique order identifier.
	ID int64 `json:"id"`

	// Ticker is the instrument ticker (e.g. "AAPL_US_EQ").
	Ticker string `json:"ticker"`

	// Type is the order type.
	Type OrderType `json:"type"`

	// Status is the current order status.
	Status OrderStatus `json:"status"`

	// Side is BUY or SELL.
	Side OrderSide `json:"side"`

	// Quantity is the total ordered quantity. Negative values represent sell
	// orders.
	Quantity float64 `json:"quantity"`

	// FilledQuantity is the quantity that has been executed so far.
	FilledQuantity float64 `json:"filledQuantity"`

	// LimitPrice is the limit price (LIMIT and STOP_LIMIT orders only).
	LimitPrice *float64 `json:"limitPrice,omitempty"`

	// StopPrice is the stop trigger price (STOP and STOP_LIMIT orders only).
	StopPrice *float64 `json:"stopPrice,omitempty"`

	// TimeValidity is how long the order remains active.
	TimeValidity TimeValidity `json:"timeValidity,omitempty"`

	// ExtendedHours indicates whether the order can execute outside regular
	// trading hours (MARKET orders only).
	ExtendedHours bool `json:"extendedHours,omitempty"`

	// DateCreated is when the order was placed.
	DateCreated time.Time `json:"dateCreated"`

	// DateModified is when the order was last modified.
	DateModified *time.Time `json:"dateModified,omitempty"`
}

// ---------------------------------------------------------------------------
// Place order request types
// ---------------------------------------------------------------------------

// MarketOrderRequest is the request body for [Client.PlaceMarketOrder].
type MarketOrderRequest struct {
	// Ticker is the instrument ticker (e.g. "AAPL_US_EQ"). Required.
	Ticker string `json:"ticker"`

	// Quantity is the number of shares to buy or sell. Use a negative value
	// for a sell order.
	Quantity float64 `json:"quantity"`

	// ExtendedHours allows the order to execute outside regular trading hours.
	ExtendedHours bool `json:"extendedHours,omitempty"`
}

// LimitOrderRequest is the request body for [Client.PlaceLimitOrder].
type LimitOrderRequest struct {
	// Ticker is the instrument ticker. Required.
	Ticker string `json:"ticker"`

	// Quantity is the number of shares. Negative for a sell order.
	Quantity float64 `json:"quantity"`

	// LimitPrice is the maximum (buy) or minimum (sell) execution price.
	// Required.
	LimitPrice float64 `json:"limitPrice"`

	// TimeValidity controls how long the order stays active. Required.
	TimeValidity TimeValidity `json:"timeValidity"`
}

// StopOrderRequest is the request body for [Client.PlaceStopOrder].
type StopOrderRequest struct {
	// Ticker is the instrument ticker. Required.
	Ticker string `json:"ticker"`

	// Quantity is the number of shares. Negative for a sell order.
	Quantity float64 `json:"quantity"`

	// StopPrice is the price that triggers the market order. Required.
	StopPrice float64 `json:"stopPrice"`

	// TimeValidity controls how long the order stays active. Required.
	TimeValidity TimeValidity `json:"timeValidity"`
}

// StopLimitOrderRequest is the request body for [Client.PlaceStopLimitOrder].
type StopLimitOrderRequest struct {
	// Ticker is the instrument ticker. Required.
	Ticker string `json:"ticker"`

	// Quantity is the number of shares. Negative for a sell order.
	Quantity float64 `json:"quantity"`

	// StopPrice is the trigger price. Required.
	StopPrice float64 `json:"stopPrice"`

	// LimitPrice is the limit price applied once the stop triggers. Required.
	LimitPrice float64 `json:"limitPrice"`

	// TimeValidity controls how long the order stays active. Required.
	TimeValidity TimeValidity `json:"timeValidity"`
}

// ---------------------------------------------------------------------------
// Order methods
// ---------------------------------------------------------------------------

// GetOrders returns all pending orders for the authenticated account.
//
// Rate limit: 1 request / 5 s.
func (c *Client) GetOrders(ctx context.Context) ([]Order, error) {
	var out []Order
	_, err := c.do(ctx, http.MethodGet, "/equity/orders", nil, &out)
	return out, err
}

// GetOrder returns the details of a specific order by its ID.
//
// Rate limit: 1 request / 1 s.
func (c *Client) GetOrder(ctx context.Context, id int64) (Order, error) {
	var out Order
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/equity/orders/%d", id), nil, &out)
	return out, err
}

// PlaceMarketOrder places a market order.
//
// Rate limit: 50 requests / 60 s.
func (c *Client) PlaceMarketOrder(ctx context.Context, req MarketOrderRequest) (Order, error) {
	var out Order
	_, err := c.do(ctx, http.MethodPost, "/equity/orders/market", req, &out)
	return out, err
}

// PlaceLimitOrder places a limit order.
//
// Rate limit: 1 request / 2 s.
func (c *Client) PlaceLimitOrder(ctx context.Context, req LimitOrderRequest) (Order, error) {
	var out Order
	_, err := c.do(ctx, http.MethodPost, "/equity/orders/limit", req, &out)
	return out, err
}

// PlaceStopOrder places a stop order.
//
// Rate limit: 1 request / 2 s.
func (c *Client) PlaceStopOrder(ctx context.Context, req StopOrderRequest) (Order, error) {
	var out Order
	_, err := c.do(ctx, http.MethodPost, "/equity/orders/stop", req, &out)
	return out, err
}

// PlaceStopLimitOrder places a stop-limit order.
//
// Rate limit: 1 request / 2 s.
func (c *Client) PlaceStopLimitOrder(ctx context.Context, req StopLimitOrderRequest) (Order, error) {
	var out Order
	_, err := c.do(ctx, http.MethodPost, "/equity/orders/stop_limit", req, &out)
	return out, err
}

// CancelOrder cancels a pending order by its ID.
//
// Rate limit: 50 requests / 60 s.
func (c *Client) CancelOrder(ctx context.Context, id int64) error {
	_, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/equity/orders/%d", id), nil, nil)
	return err
}
