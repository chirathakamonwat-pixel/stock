package gateways

import (
	"encoding/json"
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type categoryRequest struct {
	Name        *string         `json:"name"`
	Description json.RawMessage `json:"description"`
}

func (h *HTTPGateway) GetAllCategories(ctx *fiber.Ctx) error {
	categories, err := h.CategoriesService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "cannot get categories"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: categories})
}

func (h *HTTPGateway) GetCategory(ctx *fiber.Ctx) error {
	id, err := parseCategoryID(ctx.Params("id"))
	if err != nil {
		return invalidCategoryRequest(ctx)
	}

	category, err := h.CategoriesService.GetByID(id)
	if err != nil {
		return categoryErrorResponse(ctx, err, "cannot get category")
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: category})
}

func (h *HTTPGateway) CreateCategory(ctx *fiber.Ctx) error {
	var request categoryRequest
	if err := ctx.BodyParser(&request); err != nil {
		return invalidCategoryRequest(ctx)
	}
	data, valid := categoryFields(request, true)
	if !valid {
		return invalidCategoryRequest(ctx)
	}
	if err := h.CategoriesService.Create(data); err != nil {
		return categoryErrorResponse(ctx, err, "cannot create category")
	}
	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseMessage{Message: "success"})
}

func (h *HTTPGateway) UpdateCategory(ctx *fiber.Ctx) error {
	id, err := parseCategoryID(ctx.Params("id"))
	if err != nil {
		return invalidCategoryRequest(ctx)
	}

	var request categoryRequest
	if err := ctx.BodyParser(&request); err != nil {
		return invalidCategoryRequest(ctx)
	}
	data, valid := categoryFields(request, false)
	if !valid || len(data) == 0 {
		return invalidCategoryRequest(ctx)
	}
	if err := h.CategoriesService.Update(id, data); err != nil {
		return categoryErrorResponse(ctx, err, "cannot update category")
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseMessage{Message: "success"})
}

func (h *HTTPGateway) DeleteCategory(ctx *fiber.Ctx) error {
	id, err := parseCategoryID(ctx.Params("id"))
	if err != nil {
		return invalidCategoryRequest(ctx)
	}
	if err := h.CategoriesService.Delete(id); err != nil {
		return categoryErrorResponse(ctx, err, "cannot delete category")
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func categoryFields(request categoryRequest, creating bool) (map[string]interface{}, bool) {
	data := make(map[string]interface{})
	if request.Name != nil {
		if strings.TrimSpace(*request.Name) == "" {
			return nil, false
		}
		data["name"] = *request.Name
	} else if creating {
		return nil, false
	}

	if len(request.Description) > 0 {
		if string(request.Description) == "null" {
			data["description"] = nil
		} else {
			var description string
			if err := json.Unmarshal(request.Description, &description); err != nil {
				return nil, false
			}
			data["description"] = description
		}
	}
	return data, true
}

func parseCategoryID(id string) (string, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func invalidCategoryRequest(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid category request"})
}

func categoryErrorResponse(ctx *fiber.Ctx, err error, message string) error {
	if errors.Is(err, repositories.ErrCategoryNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "category not found"})
	}
	if errors.Is(err, repositories.ErrCategoryConflict) {
		return ctx.Status(fiber.StatusConflict).JSON(entities.ResponseMessage{Message: "category conflicts with existing data"})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: message})
}
