package main

import (
	"storageSys/handlers"
	"storageSys/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 允许跨域
	r.Use(middleware.Cors())

	// 路由组
	api := r.Group("/api")
	{
		// 货物相关路由
		goods := api.Group("/goods")
		{
			goods.GET("/list", handlers.GetGoodsList)
			goods.POST("/inbound", handlers.CreateInbound)
			goods.PUT("/:id", handlers.UpdateGoods)
			goods.DELETE("/:id", handlers.DeleteGoods)
			goods.POST("/outbound/:id", handlers.OutboundGoods)
			goods.POST("/mortgage/:id", handlers.MortgageGoods)
		}
	}

	r.Run(":8080")
}
