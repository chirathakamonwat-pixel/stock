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

func TestCategoriesRepositoryUsesSupabaseCRUDEndpoints(t *testing.T) {
	const categoryID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	categoryResponse := `[{"id":"` + categoryID + `","name":"Tools","description":null,"created_at":"2026-11-23T00:00:00Z"}]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/rest/v1/categories") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, categoryResponse)
		case http.MethodPost:
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create body: %v", err)
			}
			if body["name"] != "Tools" || body["description"] != "Hand tools" {
				t.Errorf("unexpected create body: %#v", body)
			}
			if r.Header.Get("Prefer") != "return=minimal" {
				t.Errorf("unexpected create Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodPatch:
			if r.URL.Query().Get("id") != "eq."+categoryID {
				t.Errorf("unexpected update filter: %s", r.URL.RawQuery)
			}
			if r.Header.Get("Prefer") != "return=representation" {
				t.Errorf("unexpected update Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, categoryResponse)
		case http.MethodDelete:
			if r.URL.Query().Get("id") != "eq."+categoryID {
				t.Errorf("unexpected delete filter: %s", r.URL.RawQuery)
			}
			if r.Header.Get("Prefer") != "return=representation" {
				t.Errorf("unexpected delete Prefer header: %q", r.Header.Get("Prefer"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, categoryResponse)
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
	repository := NewCategoriesRepository(client)

	categories, err := repository.FindAll()
	if err != nil || len(*categories) != 1 || (*categories)[0].Name != "Tools" || (*categories)[0].Description != nil {
		t.Fatalf("FindAll result = %v, error = %v", categories, err)
	}

	category, err := repository.FindByID(categoryID)
	if err != nil || category.ID != categoryID || category.Name != "Tools" {
		t.Fatalf("FindByID result = %v, error = %v", category, err)
	}

	if err := repository.Insert(map[string]interface{}{"name": "Tools", "description": "Hand tools"}); err != nil {
		t.Fatalf("Insert returned an error: %v", err)
	}
	if err := repository.Update(categoryID, map[string]interface{}{"description": nil}); err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}
	if err := repository.Delete(categoryID); err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}
}

func TestCategoriesRepositoryReturnsNotFoundForMissingCategory(t *testing.T) {
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
	repository := NewCategoriesRepository(client)

	if _, err := repository.FindByID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("FindByID error = %v, want ErrCategoryNotFound", err)
	}
	if err := repository.Update("6ba7b810-9dad-11d1-80b4-00c04fd430c8", map[string]interface{}{"name": "Missing"}); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("Update error = %v, want ErrCategoryNotFound", err)
	}
	if err := repository.Delete("6ba7b810-9dad-11d1-80b4-00c04fd430c8"); !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("Delete error = %v, want ErrCategoryNotFound", err)
	}
}
