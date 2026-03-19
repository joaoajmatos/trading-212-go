package trading212_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	trading212 "github.com/joaoajmatos/trading-212-go"
)

// newTestClient creates a Client pointing at a test server and registers
// cleanup. The server is closed automatically when the test ends.
func newTestClient(t *testing.T, handler http.HandlerFunc) *trading212.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return trading212.New("test-key",
		trading212.WithBaseURL(srv.URL),
		trading212.WithHTTPClient(srv.Client()),
	)
}

func respond(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGetAccountSummary(t *testing.T) {
	want := trading212.AccountSummary{
		ID:           12345,
		CurrencyCode: "GBP",
		Cash: trading212.Cash{
			Free:  1000.50,
			Total: 5000.00,
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("want GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "test-key" {
			t.Errorf("missing or wrong Authorization header: %s", r.Header.Get("Authorization"))
		}
		respond(t, w, http.StatusOK, want)
	})

	got, err := client.GetAccountSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: got %d, want %d", got.ID, want.ID)
	}
	if got.CurrencyCode != want.CurrencyCode {
		t.Errorf("CurrencyCode: got %s, want %s", got.CurrencyCode, want.CurrencyCode)
	}
}

func TestGetAccountSummaryError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusUnauthorized, map[string]string{
			"code":    "Unauthorized",
			"message": "invalid api key",
		})
	})

	_, err := client.GetAccountSummary(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *trading212.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *trading212.Error, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode: got %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Code != "Unauthorized" {
		t.Errorf("Code: got %q, want %q", apiErr.Code, "Unauthorized")
	}
}

func TestGetOrders(t *testing.T) {
	want := []trading212.Order{
		{
			ID:     99,
			Ticker: "AAPL_US_EQ",
			Type:   trading212.OrderTypeMarket,
			Side:   trading212.OrderSideBuy,
			Status: trading212.OrderStatusPending,
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, want)
	})

	orders, err := client.GetOrders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("got %d orders, want 1", len(orders))
	}
	if orders[0].Ticker != "AAPL_US_EQ" {
		t.Errorf("Ticker: got %s, want AAPL_US_EQ", orders[0].Ticker)
	}
}

func TestPlaceMarketOrder(t *testing.T) {
	req := trading212.MarketOrderRequest{
		Ticker:   "AAPL_US_EQ",
		Quantity: 10,
	}
	want := trading212.Order{
		ID:       42,
		Ticker:   "AAPL_US_EQ",
		Type:     trading212.OrderTypeMarket,
		Side:     trading212.OrderSideBuy,
		Status:   trading212.OrderStatusPending,
		Quantity: 10,
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("want POST, got %s", r.Method)
		}
		var got trading212.MarketOrderRequest
		_ = json.NewDecoder(r.Body).Decode(&got)
		if got.Ticker != req.Ticker {
			t.Errorf("body ticker: got %s, want %s", got.Ticker, req.Ticker)
		}
		respond(t, w, http.StatusOK, want)
	})

	order, err := client.PlaceMarketOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != want.ID {
		t.Errorf("ID: got %d, want %d", order.ID, want.ID)
	}
}

func TestCancelOrder(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("want DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	if err := client.CancelOrder(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPositions(t *testing.T) {
	now := time.Now().Truncate(time.Second).UTC()
	want := []trading212.Position{
		{
			Ticker:          "TSLA_US_EQ",
			Quantity:        5,
			AveragePrice:    200.0,
			CurrentPrice:    220.0,
			InitialFillDate: now,
		},
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, want)
	})

	positions, err := client.GetPositions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("got %d positions, want 1", len(positions))
	}
	if positions[0].Ticker != "TSLA_US_EQ" {
		t.Errorf("Ticker: got %s, want TSLA_US_EQ", positions[0].Ticker)
	}
}

func TestHistoryOrdersCursor(t *testing.T) {
	pages := []trading212.Page[trading212.HistoryOrder]{
		{
			Items:      []trading212.HistoryOrder{{ID: "order-1"}, {ID: "order-2"}},
			NextCursor: "cursor-2",
		},
		{
			Items:      []trading212.HistoryOrder{{ID: "order-3"}},
			NextCursor: "",
		},
	}

	pageIdx := 0
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if pageIdx >= len(pages) {
			t.Error("too many page requests")
			respond(t, w, http.StatusOK, trading212.Page[trading212.HistoryOrder]{})
			return
		}
		respond(t, w, http.StatusOK, pages[pageIdx])
		pageIdx++
	})

	cur := client.HistoryOrders(trading212.HistoryOrdersParams{Limit: 2})
	ctx := context.Background()

	var ids []string
	for cur.Next(ctx) {
		ids = append(ids, cur.Item().ID)
	}
	if err := cur.Err(); err != nil {
		t.Fatalf("cursor error: %v", err)
	}

	want := []string{"order-1", "order-2", "order-3"}
	if len(ids) != len(want) {
		t.Fatalf("got %d items, want %d", len(ids), len(want))
	}
	for i, id := range ids {
		if id != want[i] {
			t.Errorf("item[%d]: got %s, want %s", i, id, want[i])
		}
	}
}

func TestErrorRateLimitHeaders(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.Header().Set("X-Ratelimit-Remaining", "0")
		respond(t, w, http.StatusTooManyRequests, map[string]string{
			"code":    "TooManyRequests",
			"message": "rate limit exceeded",
		})
	})

	_, err := client.GetAccountSummary(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *trading212.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *trading212.Error, got %T", err)
	}
	if apiErr.StatusCode != 429 {
		t.Errorf("StatusCode: got %d, want 429", apiErr.StatusCode)
	}
	if apiErr.RateLimit.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter: got %v, want 5s", apiErr.RateLimit.RetryAfter)
	}
	if apiErr.RateLimit.Remaining != 0 {
		t.Errorf("Remaining: got %d, want 0", apiErr.RateLimit.Remaining)
	}
}
