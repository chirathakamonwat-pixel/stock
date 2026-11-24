package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/src/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type productsServiceStub struct {
	created map[string]interface{}
	updated map[string]interface{}
	err     error
}

func (stub *productsServiceStub) GetAll() (*[]entities.Product, error) {
	products := []entities.Product{}
	return &products, nil
}

func (stub *productsServiceStub) GetByID(string) (*entities.Product, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	return &entities.Product{ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", Name: "Widget"}, nil
}

func (stub *productsServiceStub) Create(data map[string]interface{}) error {
	stub.created = data
	return stub.err
}

func (stub *productsServiceStub) Update(_ string, data map[string]interface{}) error {
	stub.updated = data
	return stub.err
}

func (stub *productsServiceStub) Delete(string) error {
	return stub.err
}

var _ services.IProductsService = (*productsServiceStub)(nil)

func TestCreateProductUsesSchemaAndDatabaseDefaults(t *testing.T) {
	service := &productsServiceStub{}
	app := fiber.New()
	GatewayProducts(HTTPGateway{ProductsService: service}, app)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/products/create_product", strings.NewReader(`{"name":"Widget","sku":"W-1","price":12.5,"stock":0,"min_stock":0}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	if service.created["name"] != "Widget" || service.created["sku"] != "W-1" ||
		service.created["price"] != 12.5 || service.created["stock"] != 0 || service.created["min_stock"] != 0 {
		t.Fatalf("unexpected product passed to service: %#v", service.created)
	}
	if _, exists := service.created["unit"]; exists {
		t.Fatalf("unit should be omitted to use database default: %#v", service.created)
	}
}

func TestUpdateProductCanClearCategoryAndRejectsInvalidProduct(t *testing.T) {
	service := &productsServiceStub{}
	app := fiber.New()
	gateway := HTTPGateway{ProductsService: service}
	GatewayProducts(gateway, app)
	app.Patch("/api/v1/products/:id", gateway.UpdateProduct)

	request := httptest.NewRequest(http.MethodPatch, "/api/v1/products/6ba7b810-9dad-11d1-80b4-00c04fd430c8", strings.NewReader(`{"category_id":null}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	categoryID, categoryIDSet := service.updated["category_id"]
	if response.StatusCode != http.StatusOK || !categoryIDSet || categoryID != nil {
		t.Fatalf("unexpected status or update body: %d, %#v", response.StatusCode, service.updated)
	}

	invalid := httptest.NewRequest(http.MethodPost, "/api/v1/products/create_product", strings.NewReader(`{"name":"Widget","sku":"W-1","price":12.345}`))
	invalid.Header.Set("Content-Type", "application/json")
	response, err = app.Test(invalid)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected validation status: %d", response.StatusCode)
	}
}

func TestProductEndpointsReturnNotFound(t *testing.T) {
	service := &productsServiceStub{err: repositories.ErrProductNotFound}
	app := fiber.New()
	GatewayProducts(HTTPGateway{ProductsService: service}, app)

	for _, pathAndMethod := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/products/6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{http.MethodDelete, "/api/v1/products/6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
	} {
		request := httptest.NewRequest(pathAndMethod.method, pathAndMethod.path, strings.NewReader(`{"name":"Widget"}`))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNotFound {
			t.Errorf("%s %s status = %d, want %d", pathAndMethod.method, pathAndMethod.path, response.StatusCode, http.StatusNotFound)
		}
	}
}

func TestCreateProductReturnsConflictForDuplicateSKU(t *testing.T) {
	service := &productsServiceStub{err: repositories.ErrProductConflict}
	app := fiber.New()
	GatewayProducts(HTTPGateway{ProductsService: service}, app)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/products/create_product", strings.NewReader(`{"name":"Widget","sku":"W-1","price":12.5}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func TestInvalidProductIDRejected(t *testing.T) {
	service := &productsServiceStub{}
	app := fiber.New()
	GatewayProducts(HTTPGateway{ProductsService: service}, app)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/products/not-a-uuid", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
	if service.updated != nil || service.created != nil {
		t.Fatal("invalid product ID should not reach product write methods")
	}
}
