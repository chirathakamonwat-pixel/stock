package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	StockMovementsService service.IStockMovementsService
	ProductsService       service.IProductsService
	CategoriesService     service.ICategoriesService
}

func NewHTTPGateway(app *fiber.App, stockMovements service.IStockMovementsService, products service.IProductsService, categories service.ICategoriesService) {
	gateway := &HTTPGateway{
		StockMovementsService: stockMovements,
		ProductsService:       products,
		CategoriesService:     categories,
	}

	GatewayStockMovements(*gateway, app)
	GatewayProducts(*gateway, app)
	GatewayCategories(*gateway, app)
}
