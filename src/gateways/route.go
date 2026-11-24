package gateways

import "github.com/gofiber/fiber/v2"

func GatewayStockMovements(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/stock-movements")

	api.Post("/create_stock", gateway.CreateStockMovement)
	api.Get("/stock", gateway.GetAllStockMovements)
}

func GatewayProducts(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/products")

	api.Get("/products", gateway.GetAllProducts)
	api.Post("/create_product", gateway.CreateProduct)
	api.Get("/:id", gateway.GetProduct)
	//api.Patch("/:id", gateway.UpdateProduct)
	api.Delete("/:id", gateway.DeleteProduct)
}

func GatewayCategories(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/categories")

	api.Get("/categories", gateway.GetAllCategories)
	api.Post("/create_category", gateway.CreateCategory)
	api.Get("/:id", gateway.GetCategory)
	//api.Patch("/:id", gateway.UpdateCategory)
	api.Delete("/:id", gateway.DeleteCategory)
}
