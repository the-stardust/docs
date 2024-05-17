package models

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"interview/common/global"
	"interview/database"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mgoCfg = global.CONFIG.MongoDB

type DefaultStruct struct {
	Id          string `bson:"_id" json:"id"`
	CreatedTime string `bson:"created_time" json:"created_time"`
	UpdatedTime string `bson:"updated_time" json:"updated_time"`
}

func (d *DefaultStruct) PageLimit(page, size int64) (int64, int64) {
	if size == 0 {
		size = 20
	}
	if page == 0 {
		return 0, size
	}
	return (page - 1) * size, size
}

func (d *DefaultStruct) ObjectID(strHex string) primitive.ObjectID {
	objectId, _ := primitive.ObjectIDFromHex(strHex)
	return objectId
}

func (df *DefaultStruct) DefaultId() {
	if df.Id == "" {
		nowTime := time.Now().UnixNano()
		rand.Seed(time.Now().UnixNano())
		random := rand.Intn(100000)
		s := fmt.Sprintf("新途径%d-%d", nowTime, random)
		sum := md5.Sum([]byte(s))
		df.Id = hex.EncodeToString(sum[:])
	}
}
func (df *DefaultStruct) DefaultUpdateAt() {
	df.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
}

func (df *DefaultStruct) DefaultCreateAt() {
	if df.CreatedTime == "" {
		df.CreatedTime = time.Now().Format("2006-01-02 15:04:05")
	}
}
func (sf *DefaultStruct) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}

// redis pool
func (sf *DefaultStruct) RDBPool(dbName ...string) *redis.Pool {

	return global.REDISPOOL
}
func (sf *DefaultStruct) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}
func (sf *DefaultStruct) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *DefaultStruct) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}

type DefaultField struct {
	Id          primitive.ObjectID `bson:"_id" json:"id" redis:"id"`
	CreatedTime string             `bson:"created_time" json:"created_time" redis:"created_time"`
	UpdatedTime string             `bson:"updated_time" json:"updated_time" redis:"updated_time"`
}

// 1游泳圈 2多解题库
func (sf *DefaultField) GetAppSourceType(c *gin.Context) int {
	return c.GetInt("APP-SOURCE-TYPE")
}

// 用于隔离游泳圈和多解题库数据
func (sf *DefaultField) QuarantineSource(filter bson.M, c *gin.Context) bson.M {
	filter["source_type"] = c.GetInt("APP-SOURCE-TYPE")
	return filter
}
func (df *DefaultField) DefaultUpdateAt() {
	df.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
}

func (df *DefaultField) DefaultCreateAt() {
	if df.CreatedTime == "" {
		df.CreatedTime = time.Now().Format("2006-01-02 15:04:05")
	}
}

func (df *DefaultField) DefaultId() {
	if df.Id.IsZero() {
		df.Id = primitive.NewObjectID()
	}
}

func (sf *DefaultField) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}

// redis pool
func (sf *DefaultField) RDBPool(dbName ...string) *redis.Pool {
	return global.REDISPOOL
}
func (sf *DefaultField) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}

// string->primitive id
func (sf *DefaultField) ObjectID(strHex string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(strHex)
	if err != nil {
		return objectId
	}
	return objectId
}

// strings->primitive ids
func (sf *DefaultField) ObjectIDs(ids []string, noError ...bool) []primitive.ObjectID {
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

// 日志
func (sf *DefaultField) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *DefaultField) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}

type ManagerInfo struct {
	Id               string
	Name             string
	DepartmentId     string
	DepartmentName   string
	DepartmentLevel  int
	PuisneManagerIds []string `bson:"puisne_manager_ids"`
}
type SUInfo struct {
	UserId     string `json:"user_id"`
	Name       string `json:"name"` //姓名
	Phone      string `json:"phone"`
	Seat       string `json:"seat"`
	University string `json:"university"` //学校
	Specialty  string `json:"specialty"`  //专业
}
type StudentUserInfo interface {
	GetStudentUsersInfo(uids []string) map[string]SUInfo
}
