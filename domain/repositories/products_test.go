package repositories

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-fiber-template/supabase"
)

func TestProductsRepositoryUsesSupabaseCRUDEndpoints(t *testing.T) {
	const productID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	productResponse := `[{"id":"` + productID + `","name":"Widget","sku":"W-1","category_id":null,"price":12.5,"stock":0,"min_stock":5,"unit":"piece","created_at":"2026-11-23T00:00:00Z"}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/v1/products") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, productResponse)
		case http.MethodPost:
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create body: %v", err)
			}
			if body["name"] != "Widget" || body["sku"] != "W-1" || body["price"] != float64(12.5) || body["stock"] != float64(0) {
				t.Errorf("unexpected create body: %#v", body)
			}
			if r.Header.Get("Prefer") != "return=minimal" {
				t.Errorf("unexpected create Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodPatch:
			if r.URL.Query().Get("id") != "eq."+productID {
				t.Errorf("unexpected update filter: %s", r.URL.RawQuery)
			}
			if r.Header.Get("Prefer") != "return=representation" {
				t.Errorf("unexpected update Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, productResponse)
		case http.MethodDelete:
			if r.URL.Query().Get("id") != "eq."+productID {
				t.Errorf("unexpected delete filter: %s", r.URL.RawQuery)
			}
			if r.Header.Get("Prefer") != "return=representation" {
				t.Errorf("unexpected delete Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, productResponse)
		default:
			t.Errorf("unexpected HTTP method: %s", r.Method)
		}
	}))
	defer server.Close()

	t.Setenv("SUPABASE_URL", server.URL)
	t.Setenv("SUPABASE_KEY", "test-key")
	client, err := supabase.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	repository := NewProductsRepository(client)

	products, err := repository.FindAll()
	if err != nil || len(*products) != 1 || (*products)[0].SKU != "W-1" {
		t.Fatalf("FindAll result = %v, error = %v", products, err)
	}

	product, err := repository.FindByID(productID)
	if err != nil || product.ID != productID || product.Stock != 0 || product.MinStock != 5 {
		t.Fatalf("FindByID result = %v, error = %v", product, err)
	}

	if err := repository.Insert(map[string]interface{}{
		"name": "Widget", "sku": "W-1", "price": 12.5, "stock": 0,
	}); err != nil {
		t.Fatalf("Insert returned an error: %v", err)
	}

	if err := repository.Update(productID, map[string]interface{}{"name": "New widget"}); err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if err := repository.Delete(productID); err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}
}

func TestProductsRepositoryReturnsNotFoundForMissingProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[]`)
	}))
	defer server.Close()

	t.Setenv("SUPABASE_URL", server.URL)
	t.Setenv("SUPABASE_KEY", "test-key")
	client, err := supabase.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	repository := NewProductsRepository(client)

	if _, err := repository.FindByID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("FindByID error = %v, want ErrProductNotFound", err)
	}

	if err := repository.Update("6ba7b810-9dad-11d1-80b4-00c04fd430c8", map[string]interface{}{"name": "Missing"}); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("Update error = %v, want ErrProductNotFound", err)
	}

	if err := repository.Delete("6ba7b810-9dad-11d1-80b4-00c04fd430c8"); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("Delete error = %v, want ErrProductNotFound", err)
	}
}

func TestCheckProductsResponseIncludesSupabaseErrorBody(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusBadRequest,
		Status:     "400 Bad Request",
		Body:       io.NopCloser(strings.NewReader(`{"code":"23514","message":"price check constraint failed"}`)),
	}

	err := checkProductsResponse(response)
	if err == nil || !strings.Contains(err.Error(), `"code":"23514"`) {
		t.Fatalf("error = %v, want Supabase response body", err)
	}
}

func TestCheckProductsResponsePreservesConflictError(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusConflict,
		Status:     "409 Conflict",
		Body:       io.NopCloser(strings.NewReader(`{"code":"23505"}`)),
	}

	if err := checkProductsResponse(response); !errors.Is(err, ErrProductConflict) {
		t.Fatalf("error = %v, want ErrProductConflict", err)
	}
}
