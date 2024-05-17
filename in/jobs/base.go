package jobs

import (
	"interview/common/global"
	"interview/database"
	"interview/models"

	"github.com/garyburd/redigo/redis"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mgoCfg = global.CONFIG.MongoDB

type JobBase struct {
}

type Jobs struct {
	models.DefaultField `bson:",inline"`
	StartTime           string `bson:"start_time,omitempty"`
	Title               string `bson:"title"`
}

func (sf *JobBase) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}

// redis pool
func (sf *JobBase) RDBPool(dbName ...string) *redis.Pool {

	return global.REDISPOOL
}
func (sf *JobBase) ObjectID(strHex string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(strHex)
	if err != nil {
		sf.SLogger().Error(err)
	}
	return objectId
}

func (sf *JobBase) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}

// 日志
func (sf *JobBase) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *JobBase) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}

// 分页
func (sf *JobBase) PageLimit(page, size int64) (int64, int64) {
	if size == 0 {
		size = 20
	}
	if page == 0 {
		return 0, size
	}
	return (page - 1) * size, size
}
