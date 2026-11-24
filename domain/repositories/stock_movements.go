package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-fiber-template/domain/entities"
	"go-fiber-template/supabase"
	"io"
	"net/http"
	"strings"

	fiberlog "github.com/gofiber/fiber/v2/log"
)

type StockMovementInsertError struct {
	StatusCode int
	Status     string
	Code       string
	Message    string
	Details    string
	Hint       string
}

func (err *StockMovementInsertError) Error() string {
	parts := make([]string, 0, 4)
	if err.Code != "" {
		parts = append(parts, "code "+err.Code)
	}
	if err.Message != "" {
		parts = append(parts, err.Message)
	}
	if err.Details != "" {
		parts = append(parts, "details: "+err.Details)
	}
	if err.Hint != "" {
		parts = append(parts, "hint: "+err.Hint)
	}
	if len(parts) == 0 {
		parts = append(parts, "Supabase returned an error")
	}
	return fmt.Sprintf("Supabase insert stock movement returned status %s: %s", err.Status, strings.Join(parts, "; "))
}

type stockMovementsRepository struct {
	client *supabase.Client
}

type IStockMovementsRepository interface {
	Insert(data entities.StockMovement) error
	FindAll() (*[]entities.StockMovement, error)
}

func NewStockMovementsRepository(client *supabase.Client) IStockMovementsRepository {
	return &stockMovementsRepository{
		client: client,
	}
}

func (repo *stockMovementsRepository) Insert(data entities.StockMovement) error {
	body, err := json.Marshal(data)
	if err != nil {
		fiberlog.Errorf("Stock movements -> Insert: %s \n", err)
		return err
	}

	response, err := repo.client.Do(context.Background(), http.MethodPost, "stock_movements", bytes.NewReader(body), "return=minimal")
	if err != nil {
		fiberlog.Errorf("Stock movements -> Insert: %s \n", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		err = parseStockMovementInsertError(response)
		fiberlog.Errorf("Stock movements -> Insert: %s \n", err)
		return err
	}

	return nil
}

func parseStockMovementInsertError(response *http.Response) error {
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 64*1024))
	requestErr := &StockMovementInsertError{
		StatusCode: response.StatusCode,
		Status:     response.Status,
	}
	if readErr != nil {
		requestErr.Message = "failed to read Supabase error response: " + readErr.Error()
		return requestErr
	}

	var details struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details"`
		Hint    string `json:"hint"`
	}
	if err := json.Unmarshal(body, &details); err == nil {
		requestErr.Code = details.Code
		requestErr.Message = details.Message
		requestErr.Details = details.Details
		requestErr.Hint = details.Hint
	} else {
		requestErr.Message = strings.TrimSpace(string(body))
	}
	return requestErr
}

func (repo *stockMovementsRepository) FindAll() (*[]entities.StockMovement, error) {
	var movements []entities.StockMovement

	response, err := repo.client.Do(context.Background(), http.MethodGet, "stock_movements?select=*", nil, "")
	if err != nil {
		fiberlog.Errorf("Stock movements -> FindAll: %s \n", err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		err = fmt.Errorf("Supabase get stock movements returned status %s", response.Status)
		fiberlog.Errorf("Stock movements -> FindAll: %s \n", err)
		return nil, err
	}

	if err = json.NewDecoder(response.Body).Decode(&movements); err != nil {
		fiberlog.Errorf("Stock movements -> FindAll: %s \n", err)
		return nil, err
	}

	return &movements, nil
}
