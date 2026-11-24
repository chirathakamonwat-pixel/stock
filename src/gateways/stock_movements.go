package gateways

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type createStockMovementRequest struct {
	ProductID     string  `json:"product_id"`
	Type          string  `json:"type"`
	Quantity      *int    `json:"quantity"`
	PreviousStock *int    `json:"previous_stock"`
	CurrentStock  *int    `json:"current_stock"`
	Reason        *string `json:"reason"`
}

func (h *HTTPGateway) GetAllStockMovements(ctx *fiber.Ctx) error {
	data, err := h.StockMovementsService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get stock movements"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateStockMovement(ctx *fiber.Ctx) error {
	var body createStockMovementRequest
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if _, err := uuid.Parse(body.ProductID); err != nil ||
		(body.Type != "IN" && body.Type != "OUT") ||
		body.Quantity == nil || *body.Quantity <= 0 ||
		body.PreviousStock == nil || body.CurrentStock == nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	movement := entities.StockMovement{
		ProductID:     body.ProductID,
		Type:          body.Type,
		Quantity:      *body.Quantity,
		PreviousStock: *body.PreviousStock,
		CurrentStock:  *body.CurrentStock,
		Reason:        body.Reason,
	}
	if err := h.StockMovementsService.Insert(movement); err != nil {
		var insertErr *repositories.StockMovementInsertError
		if errors.As(err, &insertErr) {
			if insertErr.StatusCode == fiber.StatusConflict {
				if insertErr.Code == "23503" {
					return ctx.Status(fiber.StatusConflict).JSON(entities.ResponseMessage{Message: "product_id must reference an existing product"})
				}
				return ctx.Status(fiber.StatusConflict).JSON(entities.ResponseMessage{Message: "stock movement conflicts with a product or database constraint"})
			}
			if insertErr.StatusCode == fiber.StatusBadRequest {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "database rejected stock movement"})
			}
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot create stock movement"})
	}
	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseModel{Message: "success"})
}
