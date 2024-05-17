package middlewares

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	Middleware
}

func (sf *App) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetHeader("X-XTJ-APPID")
		c.Set("X-XTJ-APPID", appId)
		if appId == "wx3464eac698fd9ad3" {
			//面试云
			c.Set("APP-SOURCE-TYPE", 5)
			c.Set("APP-CODE", "400")
		} else if appId == "wx838f95f8c3c1ea68" {
			//面试ai
			c.Set("APP-SOURCE-TYPE", 6)
			c.Set("APP-CODE", "402")
		} else {
			c.Set("APP-SOURCE-TYPE", 0)
		}
		// 处理请求
		c.Next()
	}
}
