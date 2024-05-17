package middlewares

import (
	"interview/common"
	"interview/common/global"
	"interview/database"
	"net/http"

	"github.com/garyburd/redigo/redis"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var mgoCfg = global.CONFIG.MongoDB

type Middleware struct {
}

func (sf *Middleware) Error(code common.ResCode, c *gin.Context, msg ...string) {
	message := code.GetMsg()
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "data": make(map[string]interface{}), "message": message})
	c.Abort()
}
func (sf *Middleware) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}

// redis pool
func (sf *Middleware) RDBPool(dbName ...string) *redis.Pool {

	return global.REDISPOOL
}

// 日志
func (sf *Middleware) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *Middleware) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}
