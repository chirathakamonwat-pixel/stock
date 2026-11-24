package main

import (
	"go-fiber-template/configuration"
	repo "go-fiber-template/domain/repositories"
	gw "go-fiber-template/src/gateways"
	"go-fiber-template/src/middlewares"
	sv "go-fiber-template/src/services"
	"go-fiber-template/supabase"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New(configuration.NewFiberConfiguration())

	middlewares.Logger(app)
	app.Use(recover.New())
	app.Use(cors.New())

	supabaseClient, err := supabase.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	stockMovementsRepo := repo.NewStockMovementsRepository(supabaseClient)
	stockMovementsService := sv.NewStockMovementsService(stockMovementsRepo)
	productsRepo := repo.NewProductsRepository(supabaseClient)
	productsService := sv.NewProductsService(productsRepo)
	categoriesRepo := repo.NewCategoriesRepository(supabaseClient)
	categoriesService := sv.NewCategoriesService(categoriesRepo)

	gw.NewHTTPGateway(app, stockMovementsService, productsService, categoriesService)

	port := os.Getenv("PORT")
	if port == "" {
		port = "1818"
	}

	log.Println("Server running on port", port)
	log.Fatal(app.Listen(":" + port))
}
