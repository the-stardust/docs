package dao

import (
	"github.com/garyburd/redigo/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"interview/common/global"
	"interview/database"
)

var mgoCfg = global.CONFIG.MongoDB

type baseDao struct {
}

func (sf *baseDao) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *baseDao) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}
func (sf *baseDao) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}
func (sf *baseDao) Mysql() *gorm.DB {
	return global.Mysql
}

// redis pool
func (sf *baseDao) RDBPool(dbName ...string) *redis.Pool {

	return global.REDISPOOL
}
func (sf *baseDao) ObjectID(strHex string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(strHex)
	if err != nil {
		return objectId
	}
	return objectId
}

// strings->primitive ids
func (sf *baseDao) ObjectIDs(ids []string, noError ...bool) []primitive.ObjectID {
	objectIDs := make([]primitive.ObjectID, 0)
	for i := 0; i < len(ids); i++ {
		objectId, err := primitive.ObjectIDFromHex(ids[i])
		if err != nil {
			if len(noError) > 0 {
				continue
			}
			// sf.HttpError(err)
		}
		objectIDs = append(objectIDs, objectId)
	}
	return objectIDs
}

func (sf *baseDao) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}
