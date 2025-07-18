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

// @title        Mobile Order API
// @version      1.0
// @description  モバイルオーダー（事前注文・決済）システムのためのAPI仕様書です。
// @description  ユーザー認証、商品情報の取得、注文処理などの機能を提供します。

// @host         localhost:8080
// @BasePath     /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 認証トークンを'Bearer 'に続けて入力してください。 (例: Bearer {JWTトークン})

func main() {
	db, closer := connectDB.NewDB()
	defer closer()

	adminRepository := repositories.NewAdminRepository(db)
	userRepository := repositories.NewUserRepository(db)
	orderRepository := repositories.NewOrderRepository(db)
	shopRepository := repositories.NewShopRepository(db)
	productRepository := repositories.NewProductRepository(db)

	adminService := services.NewAdminService(adminRepository)
	authService := services.NewAuthService(userRepository, shopRepository, orderRepository)
	orderService := services.NewOrderService(orderRepository, productRepository)
	productService := services.NewProductService(productRepository)

	adminController := controllers.NewAdminController(adminService)
	authController := controllers.NewAuthController(authService)
	orderController := controllers.NewOrderController(orderService)
	productController := controllers.NewProductController(productService)

	e := api.NewRouter(adminController, authController, orderController, productController)

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
