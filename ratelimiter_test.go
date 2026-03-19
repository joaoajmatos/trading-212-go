package trading212_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	trading212 "github.com/joaoajmatos/trading-212-go"
)

// newRLTestClient is like newTestClient but also enables WithRateLimiting.
func newRLTestClient(t *testing.T, handler http.HandlerFunc) *trading212.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return trading212.New("test-key",
		trading212.WithBaseURL(srv.URL),
		trading212.WithHTTPClient(srv.Client()),
		trading212.WithRateLimiting(),
	)
}

// TestWithRateLimiting_ThrottlesRequests verifies that WithRateLimiting spaces
// out requests that exceed the bucket capacity.
//
// The /equity/positions endpoint is limited to 1 req/s (burst = 1). Making 3
// consecutive calls must take at least 2 s: the first call consumes the initial
// token immediately; each subsequent call waits ~1 s for a refill.
func TestWithRateLimiting_ThrottlesRequests(t *testing.T) {
	client := newRLTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, []trading212.Position{})
	})

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 3; i++ {
		if _, err := client.GetPositions(ctx); err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	if elapsed < 2*time.Second {
		t.Errorf("expected ≥ 2s for 3 calls at 1 req/s, got %v", elapsed)
	}
}

// TestWithRateLimiting_ContextCancellation verifies that a call waiting for a
// rate-limit slot returns immediately when the context is cancelled.
func TestWithRateLimiting_ContextCancellation(t *testing.T) {
	client := newRLTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, []trading212.Position{})
	})

	ctx := context.Background()

	// Drain the bucket (burst = 1).
	if _, err := client.GetPositions(ctx); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Bucket is now empty. Cancel the context before the next call so the
	// wait is aborted immediately.
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	_, err := client.GetPositions(cancelCtx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestWithRateLimiting_BurstAllowed verifies that up to burst-many requests
// succeed immediately without waiting.
//
// The market order endpoint allows 50 req/60 s. All 50 should complete
// without any throttling delay.
func TestWithRateLimiting_BurstAllowed(t *testing.T) {
	client := newRLTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(t, w, http.StatusOK, trading212.Order{})
	})

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 50; i++ {
		_, err := client.PlaceMarketOrder(ctx, trading212.MarketOrderRequest{
			Ticker:   "AAPL_US_EQ",
			Quantity: 1,
		})
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// All 50 tokens were available upfront; no waiting should occur.
	if elapsed > 2*time.Second {
		t.Errorf("burst of 50 requests took %v, expected < 2s", elapsed)
	}
}
