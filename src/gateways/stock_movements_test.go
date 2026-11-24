package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/src/services"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type stockMovementsServiceStub struct {
	inserted     entities.StockMovement
	insertCalled bool
	insertErr    error
}

func (stub *stockMovementsServiceStub) GetAll() (*[]entities.StockMovement, error) {
	movements := []entities.StockMovement{}
	return &movements, nil
}

func (stub *stockMovementsServiceStub) Insert(movement entities.StockMovement) error {
	stub.inserted = movement
	stub.insertCalled = true
	return stub.insertErr
}

var _ services.IStockMovementsService = (*stockMovementsServiceStub)(nil)

func TestCreateStockMovementValidatesAndInsertsSchemaFields(t *testing.T) {
	service := &stockMovementsServiceStub{}
	app := fiber.New()
	GatewayStockMovements(HTTPGateway{StockMovementsService: service}, app)

	body := `{"product_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","type":"IN","quantity":5,"previous_stock":2,"current_stock":7,"reason":"restock"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/stock-movements/create_stock", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	if !service.insertCalled ||
		service.inserted.ProductID != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" ||
		service.inserted.Type != "IN" ||
		service.inserted.Quantity != 5 ||
		service.inserted.PreviousStock != 2 ||
		service.inserted.CurrentStock != 7 ||
		service.inserted.Reason == nil ||
		*service.inserted.Reason != "restock" ||
		service.inserted.ID != "" ||
		service.inserted.CreatedAt != nil {
		t.Fatalf("unexpected inserted movement: %+v", service.inserted)
	}
}

func TestCreateStockMovementRejectsInvalidSchemaFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid product id",
			body: `{"product_id":"not-a-uuid","type":"IN","quantity":1,"previous_stock":0,"current_stock":1}`,
		},
		{
			name: "invalid type",
			body: `{"product_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","type":"ADJUST","quantity":1,"previous_stock":0,"current_stock":1}`,
		},
		{
			name: "non-positive quantity",
			body: `{"product_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","type":"OUT","quantity":0,"previous_stock":1,"current_stock":1}`,
		},
		{
			name: "missing stock values",
			body: `{"product_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","type":"IN","quantity":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &stockMovementsServiceStub{}
			app := fiber.New()
			GatewayStockMovements(HTTPGateway{StockMovementsService: service}, app)

			request := httptest.NewRequest(http.MethodPost, "/api/v1/stock-movements/create_stock", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusUnprocessableEntity {
				t.Fatalf("unexpected status: %d", response.StatusCode)
			}
			if service.insertCalled {
				t.Fatal("invalid stock movement was passed to the service")
			}
		})
	}
}

func TestCreateStockMovementReturnsConflictFromSupabase(t *testing.T) {
	service := &stockMovementsServiceStub{
		insertErr: &repositories.StockMovementInsertError{
			StatusCode: http.StatusConflict,
			Status:     "409 Conflict",
			Code:       "23503",
			Message:    "foreign key violation",
		},
	}
	app := fiber.New()
	GatewayStockMovements(HTTPGateway{StockMovementsService: service}, app)

	body := `{"product_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","type":"IN","quantity":1,"previous_stock":0,"current_stock":1}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/stock-movements/create_stock", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(responseBody), "product_id must reference an existing product") {
		t.Fatalf("unexpected conflict response: %s", responseBody)
	}
}
