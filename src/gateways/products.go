package gateways

import (
	"encoding/json"
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"math"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type productRequest struct {
	Name       *string         `json:"name"`
	SKU        *string         `json:"sku"`
	CategoryID json.RawMessage `json:"category_id"`
	Price      *float64        `json:"price"`
	Stock      *int            `json:"stock"`
	MinStock   *int            `json:"min_stock"`
	Unit       *string         `json:"unit"`
}

func (h *HTTPGateway) GetAllProducts(ctx *fiber.Ctx) error {
	products, err := h.ProductsService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot get products"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: products})
}

func (h *HTTPGateway) GetProduct(ctx *fiber.Ctx) error {
	id, err := parseProductID(ctx.Params("id"))
	if err != nil {
		return invalidProductRequest(ctx)
	}

	product, err := h.ProductsService.GetByID(id)
	if err != nil {
		return productErrorResponse(ctx, err, "cannot get product")
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: product})
}

func (h *HTTPGateway) CreateProduct(ctx *fiber.Ctx) error {
	var request productRequest
	if err := ctx.BodyParser(&request); err != nil {
		return invalidProductRequest(ctx)
	}

	data, valid := productFields(request, true)
	if !valid {
		return invalidProductRequest(ctx)
	}
	if err := h.ProductsService.Create(data); err != nil {
		return productErrorResponse(ctx, err, "cannot create product")
	}
	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseMessage{Message: "success"})
}

func (h *HTTPGateway) UpdateProduct(ctx *fiber.Ctx) error {
	id, err := parseProductID(ctx.Params("id"))
	if err != nil {
		return invalidProductRequest(ctx)
	}

	var request productRequest
	if err := ctx.BodyParser(&request); err != nil {
		return invalidProductRequest(ctx)
	}
	data, valid := productFields(request, false)
	if !valid || len(data) == 0 {
		return invalidProductRequest(ctx)
	}
	if err := h.ProductsService.Update(id, data); err != nil {
		return productErrorResponse(ctx, err, "cannot update product")
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{Message: "success"})
}

func (h *HTTPGateway) DeleteProduct(ctx *fiber.Ctx) error {
	id, err := parseProductID(ctx.Params("id"))
	if err != nil {
		return invalidProductRequest(ctx)
	}
	if err := h.ProductsService.Delete(id); err != nil {
		return productErrorResponse(ctx, err, "cannot delete product")
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func productFields(request productRequest, creating bool) (map[string]interface{}, bool) {
	data := make(map[string]interface{})

	if request.Name != nil {
		if strings.TrimSpace(*request.Name) == "" {
			return nil, false
		}
		data["name"] = *request.Name
	} else if creating {
		return nil, false
	}

	if request.SKU != nil {
		if strings.TrimSpace(*request.SKU) == "" {
			return nil, false
		}
		data["sku"] = *request.SKU
	} else if creating {
		return nil, false
	}

	if len(request.CategoryID) > 0 {
		if string(request.CategoryID) == "null" {
			data["category_id"] = nil
		} else {
			var categoryID string
			if err := json.Unmarshal(request.CategoryID, &categoryID); err != nil {
				return nil, false
			}
			if _, err := uuid.Parse(categoryID); err != nil {
				return nil, false
			}
			data["category_id"] = categoryID
		}
	}

	if request.Price != nil {
		if !validPrice(*request.Price) {
			return nil, false
		}
		data["price"] = *request.Price
	} else if creating {
		return nil, false
	}

	if request.Stock != nil {
		if *request.Stock < 0 {
			return nil, false
		}
		data["stock"] = *request.Stock
	}

	if request.MinStock != nil {
		if *request.MinStock < 0 {
			return nil, false
		}
		data["min_stock"] = *request.MinStock
	}

	if request.Unit != nil {
		if strings.TrimSpace(*request.Unit) == "" {
			return nil, false
		}
		data["unit"] = *request.Unit
	}

	return data, true
}

func validPrice(price float64) bool {
	return !math.IsNaN(price) && !math.IsInf(price, 0) &&
		price >= 0 && price <= 99999999.99 &&
		math.Abs(price*100-math.Round(price*100)) < 0.000001
}

func parseProductID(id string) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func invalidProductRequest(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid product request"})
}

func productErrorResponse(ctx *fiber.Ctx, err error, message string) error {
	if errors.Is(err, repositories.ErrProductNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "product not found"})
	}
	if errors.Is(err, repositories.ErrProductConflict) {
		return ctx.Status(fiber.StatusConflict).JSON(entities.ResponseMessage{Message: "product conflicts with existing data"})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: message})
}
