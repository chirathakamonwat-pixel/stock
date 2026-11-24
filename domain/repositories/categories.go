package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"go-fiber-template/domain/entities"
	"go-fiber-template/supabase"

	fiberlog "github.com/gofiber/fiber/v2/log"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryConflict = errors.New("category conflicts with existing data")

type categoriesRepository struct {
	client *supabase.Client
}

type ICategoriesRepository interface {
	FindAll() (*[]entities.Category, error)
	FindByID(id string) (*entities.Category, error)
	Insert(data map[string]interface{}) error
	Update(id string, data map[string]interface{}) error
	Delete(id string) error
}

func NewCategoriesRepository(client *supabase.Client) ICategoriesRepository {
	return &categoriesRepository{client: client}
}

func (repo *categoriesRepository) FindAll() (*[]entities.Category, error) {
	response, err := repo.client.Do(context.Background(), http.MethodGet, "categories?select=*", nil, "")
	if err != nil {
		fiberlog.Errorf("Categories -> FindAll: %s \n", err)
		return nil, err
	}
	defer response.Body.Close()

	if err := checkCategoriesResponse(response); err != nil {
		fiberlog.Errorf("Categories -> FindAll: %s \n", err)
		return nil, err
	}

	var categories []entities.Category
	if err := json.NewDecoder(response.Body).Decode(&categories); err != nil {
		fiberlog.Errorf("Categories -> FindAll: %s \n", err)
		return nil, err
	}
	return &categories, nil
}

func (repo *categoriesRepository) FindByID(id string) (*entities.Category, error) {
	response, err := repo.client.Do(context.Background(), http.MethodGet, categoryEndpoint(id), nil, "")
	if err != nil {
		fiberlog.Errorf("Categories -> FindByID: %s \n", err)
		return nil, err
	}
	defer response.Body.Close()

	if err := checkCategoriesResponse(response); err != nil {
		fiberlog.Errorf("Categories -> FindByID: %s \n", err)
		return nil, err
	}

	var categories []entities.Category
	if err := json.NewDecoder(response.Body).Decode(&categories); err != nil {
		fiberlog.Errorf("Categories -> FindByID: %s \n", err)
		return nil, err
	}
	if len(categories) == 0 {
		return nil, ErrCategoryNotFound
	}
	return &categories[0], nil
}

func (repo *categoriesRepository) Insert(data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		fiberlog.Errorf("Categories -> Insert: %s \n", err)
		return err
	}

	response, err := repo.client.Do(context.Background(), http.MethodPost, "categories", bytes.NewReader(body), "return=minimal")
	if err != nil {
		fiberlog.Errorf("Categories -> Insert: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkCategoriesResponse(response); err != nil {
		fiberlog.Errorf("Categories -> Insert: %s \n", err)
		return err
	}
	return nil
}

func (repo *categoriesRepository) Update(id string, data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		fiberlog.Errorf("Categories -> Update: %s \n", err)
		return err
	}

	response, err := repo.client.Do(context.Background(), http.MethodPatch, categoryEndpoint(id), bytes.NewReader(body), "return=representation")
	if err != nil {
		fiberlog.Errorf("Categories -> Update: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkCategoriesResponse(response); err != nil {
		fiberlog.Errorf("Categories -> Update: %s \n", err)
		return err
	}

	var updated []entities.Category
	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		fiberlog.Errorf("Categories -> Update: %s \n", err)
		return err
	}
	if len(updated) == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (repo *categoriesRepository) Delete(id string) error {
	response, err := repo.client.Do(context.Background(), http.MethodDelete, categoryEndpoint(id), nil, "return=representation")
	if err != nil {
		fiberlog.Errorf("Categories -> Delete: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if err := checkCategoriesResponse(response); err != nil {
		fiberlog.Errorf("Categories -> Delete: %s \n", err)
		return err
	}

	var deleted []entities.Category
	if err := json.NewDecoder(response.Body).Decode(&deleted); err != nil {
		fiberlog.Errorf("Categories -> Delete: %s \n", err)
		return err
	}
	if len(deleted) == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func categoryEndpoint(id string) string {
	return "categories?id=eq." + url.QueryEscape(id) + "&select=*"
}

func checkCategoriesResponse(response *http.Response) error {
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode == http.StatusConflict {
			return ErrCategoryConflict
		}
		return fmt.Errorf("Supabase categories request returned status %s", response.Status)
	}
	return nil
}
