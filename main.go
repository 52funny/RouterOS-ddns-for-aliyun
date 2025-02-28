package main

import (
	"aliyun_ddns/controllers"
	"aliyun_ddns/middlewares"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode) // 设置 Gin 为发布模式
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.Logger())
	r.GET("/aliyun_ddns", controllers.AddUpdateAliddns)
	r.GET("/ahu_dhcp", controllers.AhuDchpAddressAuth)
	r.Run(":3000") // 确保服务监听 3000 端口
}
