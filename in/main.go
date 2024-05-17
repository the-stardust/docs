package main

import (
	_ "interview/common/config"
	"interview/es"
	"interview/logger"
	"interview/router"
)

func main() {
	logger.InitZapLogger()
	r := router.InitRouter()
	es.InitElastic()
	if err := r.Run("0.0.0.0:8040"); err != nil {
		panic("startup service failed, err:" + err.Error())
	}

}
