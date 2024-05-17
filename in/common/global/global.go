package global

import (
	"interview/config"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 全局变量
var (
	VIPER         *viper.Viper
	CONFIG        *config.Server
	LOGGER        *zap.Logger
	SUGARLOGGER   *zap.SugaredLogger
	REDISDB       redis.Conn
	REDISPOOL     *redis.Pool
	REDISPOOLBank *redis.Pool
	Mysql         *gorm.DB
)
