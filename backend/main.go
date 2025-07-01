package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/A4-dev-team/mobileorder.git/api"
	"github.com/A4-dev-team/mobileorder.git/connectDB"
	"github.com/A4-dev-team/mobileorder.git/controllers"
	"github.com/A4-dev-team/mobileorder.git/repositories"
	"github.com/A4-dev-team/mobileorder.git/services"

	_ "github.com/lib/pq"
)

//	@title			mobile_order
//	@version		1.0
//	@description	This is mobileorder API
//	@host			localhost:8080
//	@BasePath		/
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "認証トークンを'Bearer 'に続けて入力してください。例: Bearer {JWTトークン}"

func main() {
	db, closer := connectDB.NewDB()
	defer closer()

	adminRepository := repositories.NewAdminRepository(db)
	userRepository := repositories.NewUserRepository(db)
	orderRepository := repositories.NewOrderRepository(db)
	shopRepository := repositories.NewShopRepository(db)

	adminService := services.NewAdminService(adminRepository)
	authService := services.NewAuthService(userRepository, shopRepository, orderRepository)
	orderService := services.NewOrderService(orderRepository)

	adminController := controllers.NewAdminController(adminService)
	authController := controllers.NewAuthController(authService)
	orderController := controllers.NewOrderController(orderService)

	e := api.NewRouter(adminController, authController, orderController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		log.Printf("server start at port %s", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting ndown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	log.Println("Server gracefully stopped")

}
