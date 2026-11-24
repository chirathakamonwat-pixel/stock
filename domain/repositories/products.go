package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go-fiber-template/domain/entities"
	"go-fiber-template/supabase"

	fiberlog "github.com/gofiber/fiber/v2/log"
)

var ErrProductNotFound = errors.New("product not found")
var ErrProductConflict = errors.New("product conflicts with existing data")

type productsRepository struct {
	client *supabase.Client
}

type IProductsRepository interface {
	FindAll() (*[]entities.Product, error)
	FindByID(id string) (*entities.Product, error)
	Insert(data map[string]interface{}) error
	Update(id string, data map[string]interface{}) error
	Delete(id string) error
}

func NewProductsRepository(client *supabase.Client) IProductsRepository {
	return &productsRepository{client: client}
}

func (repo *productsRepository) FindAll() (*[]entities.Product, error) {
	response, err := repo.client.Do(context.Background(), http.MethodGet, "products?select=*", nil, "")
	if err != nil {
		fiberlog.Errorf("Products -> FindAll: %s \n", err)
		return nil, err
	}
	defer response.Body.Close()

	if err := checkProductsResponse(response); err != nil {
		fiberlog.Errorf("Products -> FindAll: %s \n", err)
		return nil, err
	}

	var products []entities.Product
	if err := json.NewDecoder(response.Body).Decode(&products); err != nil {
		fiberlog.Errorf("Products -> FindAll: %s \n", err)
		return nil, err
	}
	return &products, nil
}

func (repo *productsRepository) FindByID(id string) (*entities.Product, error) {
	response, err := repo.client.Do(context.Background(), http.MethodGet, productEndpoint(id), nil, "")
	if err != nil {
		fiberlog.Errorf("Products -> FindByID: %s \n", err)
		return nil, err
	}
	defer response.Body.Close()

	if err := checkProductsResponse(response); err != nil {
		fiberlog.Errorf("Products -> FindByID: %s \n", err)
		return nil, err
	}

	var products []entities.Product
	if err := json.NewDecoder(response.Body).Decode(&products); err != nil {
		fiberlog.Errorf("Products -> FindByID: %s \n", err)
		return nil, err
	}
	if len(products) == 0 {
		return nil, ErrProductNotFound
	}
	return &products[0], nil
}

func (repo *productsRepository) Update(id string, data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		fiberlog.Errorf("Products -> Update: %s \n", err)
		return err
	}

	response, err := repo.client.Do(context.Background(), http.MethodPatch, productEndpoint(id), bytes.NewReader(body), "return=representation")
	if err != nil {
		fiberlog.Errorf("Products -> Update: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkProductsResponse(response); err != nil {
		fiberlog.Errorf("Products -> Update: %s \n", err)
		return err
	}

	var updated []entities.Product
	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		fiberlog.Errorf("Products -> Update: %s \n", err)
		return err
	}
	if len(updated) == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (repo *productsRepository) Delete(id string) error {
	response, err := repo.client.Do(context.Background(), http.MethodDelete, productEndpoint(id), nil, "return=representation")
	if err != nil {
		fiberlog.Errorf("Products -> Delete: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkProductsResponse(response); err != nil {
		fiberlog.Errorf("Products -> Delete: %s \n", err)
		return err
	}

	var deleted []entities.Product
	if err := json.NewDecoder(response.Body).Decode(&deleted); err != nil {
		fiberlog.Errorf("Products -> Delete: %s \n", err)
		return err
	}
	if len(deleted) == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (repo *productsRepository) Insert(data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		fiberlog.Errorf("Products -> Insert: %s \n", err)
		return err
	}

	response, err := repo.client.Do(context.Background(), http.MethodPost, "products", bytes.NewReader(body), "return=minimal")
	if err != nil {
		fiberlog.Errorf("Products -> Insert: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkProductsResponse(response); err != nil {
		fiberlog.Errorf("Products -> Insert: %s \n", err)
		return err
	}
	return nil
}

func productEndpoint(id string) string {
	return "products?id=eq." + url.QueryEscape(id) + "&select=*"
}

func checkProductsResponse(response *http.Response) error {
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode == http.StatusConflict {
			return ErrProductConflict
		}
		var errorBody string
		if response.Body != nil {
			body, err := io.ReadAll(io.LimitReader(response.Body, 64*1024))
			if err != nil {
				return fmt.Errorf("Supabase products request returned status %s; failed to read error body: %w", response.Status, err)
			}
			errorBody = strings.TrimSpace(string(body))
		}
		if errorBody != "" {
			return fmt.Errorf("Supabase products request returned status %s: %s", response.Status, errorBody)
		}
		return fmt.Errorf("Supabase products request returned status %s", response.Status)
	}
	return nil
}
