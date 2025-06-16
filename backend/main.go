package main

import (
	"log"
	"os"

	"github.com/A4-dev-team/mobileorder.git/api"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/db"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"

	_ "github.com/lib/pq"
)

func main() {
	db := db.NewDB()

	adminRepository := repositories.NewAdminRepository(db)
	authRepository := repositories.NewAuthRepository(db)
	orderRepository := repositories.NewOrderRepository(db)

	adminService := services.NewAdminService(adminRepository)
	authService := services.NewAuthService(authRepository)
	orderService := services.NewOrderService(orderRepository)

	adminController := controllers.NewAdminController(adminService)
	authController := controllers.NewAuthController(authService)
	orderController := controllers.NewOrderController(orderService)

	e := api.NewRouter(adminController, authController, orderController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	log.Printf("server start at port %s", port)
	log.Fatal(e.Start(":" + port))
}
