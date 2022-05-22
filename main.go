package main

import (
	"myfreax/v2raym/v2ray"

	"github.com/gin-gonic/gin"
)

type Error struct {
	msg string
}

func main() {
	config, _ := v2ray.Create("./v2ray.json")
	service := v2ray.Start(config)
	v2 := v2ray.V2ray{Config: config, Service: service}

	server := gin.Default()

	//all
	server.GET("/clients", func(ctx *gin.Context) {
		ctx.JSON(200, v2.Config.QueryAllClients())
	})

	//removed
	server.GET("/clients/deleted", func(ctx *gin.Context) {
		ctx.JSON(200, v2.Config.QueryRemovedClients())
	})

	// add
	server.PUT("/clients", func(ctx *gin.Context) {
		var addClient v2ray.AddClient
		if err := ctx.BindJSON(&addClient); err != nil {
			err := Error{msg: err.Error()}
			ctx.JSON(500, err)
		}
		client := v2.Config.AddClient(addClient)
		ctx.JSON(200, client)
	})

	// remove
	server.DELETE("/clients/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		v2.Config.DisableClient(id)
	})

	// enable
	server.POST("/clients", func(ctx *gin.Context) {
		var updateClient v2ray.UpdateClient
		if err := ctx.BindJSON(&updateClient); err != nil {
			err := Error{msg: err.Error()}
			ctx.JSON(500, err)
		}
		v2.Config.EnableClient(updateClient)
	})

	server.Run("127.0.0.1:8318")
}
