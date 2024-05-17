package main

import (
	_ "interview/common/config"
	_ "interview/database"
	"interview/jobs"
	"interview/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	initInMain()
	logger.InitZapLogger()
	go jobs.Run()
	go jobs.OnceRun()
	r := gin.Default()
	r.GET("/interview-job/healthy", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"K8s": "healthy",
		})
	})
	r.Run(":8041")
}

func initInMain() {
	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone
}
