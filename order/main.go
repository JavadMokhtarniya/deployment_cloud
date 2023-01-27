package main

import (
	"flag"

	"github.com/gin-gonic/gin"
)

func main() {
	authAddr := flag.String("authAddr", "localhost:50051", "Authentication service address (GRPC)")
	prodAddr := flag.String("prodAddr", "localhost:50052", "Product service address (GRPC)")

	flag.Parse()

	svc := NewInmemservice(*prodAddr)
	handler := NewHandler(svc)
	authMW := NewAuthMiddleware(*authAddr)

	router := gin.Default()

	router.GET("/api/basket", authMW.hasAccess, handler.getBasket)
	router.POST("/api/basket/:pid", authMW.hasAccess, handler.addProductToBasket)
	router.PUT("/api/basket/:pid", authMW.hasAccess, handler.modifyBasket)

	router.Run(":3000")
}
