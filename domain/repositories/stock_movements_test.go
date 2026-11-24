package repositories

import (
	"encoding/json"
	"errors"
	"go-fiber-template/domain/entities"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-fiber-template/supabase"
)

func newTestStockMovementsRepository(t *testing.T, handler http.HandlerFunc) IStockMovementsRepository {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Setenv("SUPABASE_URL", server.URL)
	t.Setenv("SUPABASE_KEY", "test-key")

	client, err := supabase.NewClient()
	if err != nil {
		server.Close()
		t.Fatal(err)
	}

	t.Cleanup(server.Close)
	return NewStockMovementsRepository(client)
}

func TestInsertStockMovementPostsToSupabase(t *testing.T) {
	repository := newTestStockMovementsRepository(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/v1/stock_movements" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("apikey") != "test-key" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Supabase authentication headers were not set")
		}
		if r.Header.Get("Prefer") != "return=minimal" {
			t.Errorf("unexpected Prefer header: %q", r.Header.Get("Prefer"))
		}

		var movement entities.StockMovement
		if err := json.NewDecoder(r.Body).Decode(&movement); err != nil {
			t.Errorf("decode stock movement request: %v", err)
		}
		if movement.ProductID != "product-1" || movement.Type != "IN" ||
			movement.Quantity != 10 || movement.PreviousStock != 2 || movement.CurrentStock != 12 {
			t.Errorf("unexpected stock movement payload: %+v", movement)
		}
		if movement.ID != "" || movement.CreatedAt != nil {
			t.Errorf("database-generated fields should not be sent: %+v", movement)
		}
		w.WriteHeader(http.StatusCreated)
	})

	err := repository.Insert(entities.StockMovement{
		ProductID:     "product-1",
		Type:          "IN",
		Quantity:      10,
		PreviousStock: 2,
		CurrentStock:  12,
	})
	if err != nil {
		t.Fatalf("Insert returned an error: %v", err)
	}
}

func TestFindAllReadsStockMovementsFromSupabase(t *testing.T) {
	repository := newTestStockMovementsRepository(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/rest/v1/stock_movements" || r.URL.Query().Get("select") != "*" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"movement-1","product_id":"product-1","type":"OUT","quantity":3,"previous_stock":12,"current_stock":9,"reason":"sale","created_at":"2026-11-23T00:00:00Z"}]`))
	})

	movements, err := repository.FindAll()
	if err != nil {
		t.Fatalf("FindAll returned an error: %v", err)
	}
	if len(*movements) != 1 || (*movements)[0].ProductID != "product-1" ||
		(*movements)[0].Type != "OUT" || (*movements)[0].Reason == nil ||
		*(*movements)[0].Reason != "sale" || (*movements)[0].CreatedAt == nil ||
		!(*movements)[0].CreatedAt.Equal(time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected stock movements: %+v", *movements)
	}
}

func TestFindAllReturnsErrorForSupabaseFailure(t *testing.T) {
	repository := newTestStockMovementsRepository(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
	})

	if _, err := repository.FindAll(); err == nil {
		t.Fatal("expected an error for a failed Supabase request")
	}
}

func TestInsertIncludesSupabaseConstraintDetails(t *testing.T) {
	repository := newTestStockMovementsRepository(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code":"23503","message":"foreign key violation","details":"product is missing","hint":"create the product first"}`))
	})

	err := repository.Insert(entities.StockMovement{
		ProductID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Type:          "IN",
		Quantity:      1,
		PreviousStock: 0,
		CurrentStock:  1,
	})
	var insertErr *StockMovementInsertError
	if !errors.As(err, &insertErr) {
		t.Fatalf("Insert error type = %T, want *StockMovementInsertError", err)
	}
	if insertErr.StatusCode != http.StatusConflict ||
		insertErr.Code != "23503" ||
		insertErr.Message != "foreign key violation" ||
		insertErr.Details != "product is missing" ||
		insertErr.Hint != "create the product first" {
		t.Fatalf("unexpected Supabase error details: %+v", insertErr)
	}
}
