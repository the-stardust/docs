package controllers

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/database"
	"interview/services"
	"math"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/olahol/melody"

	"github.com/go-playground/validator/v10"

	"github.com/garyburd/redigo/redis"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var mgoCfg = global.CONFIG.MongoDB

type Controller struct {
}

func (sf *Controller) GetAppSourceType(c *gin.Context) int {
	return c.GetInt("APP-SOURCE-TYPE")
}

// 数据库
func (sf *Controller) DB(dbName ...string) *database.MongoWork {
	db := mgoCfg.Dbname
	if len(dbName) > 0 {
		db = dbName[0]
	}
	return database.NewMongoWork(mgoCfg.Path, mgoCfg.Username, mgoCfg.Password, db)
}

// redis pool
func (sf *Controller) RDBPool(dbName ...string) *redis.Pool {
	return global.REDISPOOL
}

// success respone
func (sf *Controller) Success(data interface{}, c *gin.Context, msg ...string) {
	message := common.CodeSuccess.GetMsg()
	if len(msg) > 0 {
		message = msg[0]
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	if v, ok := data.(string); ok {
		if v == "" {
			data = make(map[string]interface{})
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "message": message})
	c.Abort()
}

// error respone
func (sf *Controller) Error(code common.ResCode, c *gin.Context, msg ...string) {
	message := code.GetMsg()
	if len(msg) > 0 {
		message = msg[0]
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "data": make(map[string]interface{}), "message": message})
	c.Abort()
}

// error respone
func (sf *Controller) SocketError(code common.ResCode, c *melody.Session, msg ...string) {
	message := code.GetMsg()
	if len(msg) > 0 {
		message = msg[0]
	}
	resp := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	r, err := json.Marshal(resp)
	if err != nil {
		sf.SLogger().Error(err)
	}
	c.Write(r)
}

// success respone
func (sf *Controller) SocketSuccess(data interface{}, c *melody.Session, msg ...string) {
	message := common.CodeSuccess.GetMsg()
	if len(msg) > 0 {
		message = msg[0]
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	if v, ok := data.(string); ok {
		if v == "" {
			data = make(map[string]interface{})
		}
	}
	resp := map[string]interface{}{
		"code": 0,
		"data": data, "message": message,
	}
	r, err := json.Marshal(resp)
	if err != nil {
		sf.SLogger().Error(err)
	}
	c.Write(r)
}

// 验证mongodb 结果
func (sf *Controller) MongoNoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}

// string->primitive id
func (sf *Controller) ObjectID(strHex string) primitive.ObjectID {
	objectId, err := primitive.ObjectIDFromHex(strHex)
	if err != nil {
		sf.SLogger().Error(err)
	}
	return objectId
}

// strings->primitive ids
func (sf *Controller) ObjectIDs(ids []string, noError ...bool) []primitive.ObjectID {
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

// 分页
func (sf *Controller) PageLimit(page, size int64) (int64, int64) {
	if size == 0 {
		size = 20
	}
	if page == 0 {
		return 0, size
	}
	return (page - 1) * size, size
}

// 切片分页
func (sf *Controller) SlicePage(page, pageSize, nums int64) (sliceStart, sliceEnd int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > nums {
		if page == 1 {
			return 0, nums
		} else {
			return 0, 0
		}

	}
	// 总页数
	pageCount := int64(math.Ceil(float64(nums) / float64(pageSize)))
	if page > pageCount {
		return 0, 0
	}
	sliceStart = (page - 1) * pageSize
	sliceEnd = sliceStart + pageSize
	if sliceEnd > nums {
		sliceEnd = nums
	}
	return sliceStart, sliceEnd
}

// 日志
func (sf *Controller) Logger() *zap.Logger {

	return global.LOGGER
}
func (sf *Controller) SLogger() *zap.SugaredLogger {

	return global.SUGARLOGGER
}

// 浮点数科学计数
func (sf *Controller) TransitionFloat64(f float64, digit int8) float64 {

	if math.IsInf(f, 1) || math.IsInf(f, -1) || math.IsNaN(f) {
		f = 0
	}
	digitStr := "%f"
	if digit >= 0 {
		digitStr = fmt.Sprintf("%%.%df", digit)
	}
	f, err := strconv.ParseFloat(fmt.Sprintf(digitStr, f), 64)
	if err != nil {
		sf.SLogger().Error(err)
		return 0
	}
	return f
}

// recover 堆栈错误信息
func (sf *Controller) RecoverStackError() {
	errs := recover()
	if errs != nil {
		var stackBuf [1024]byte
		stackBufLen := runtime.Stack(stackBuf[:], false)
		sf.SLogger().Error(string(stackBuf[:stackBufLen]))
	}
}

func (sf *Controller) GetValidMsg(err error, obj interface{}) string {
	if reflect.TypeOf(obj).Kind().String() == "ptr" {
		getObj := reflect.TypeOf(obj)
		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {
				if f, exist := getObj.Elem().FieldByName(e.Field()); exist {
					return f.Tag.Get("msg")
				}
			}
		}
		return err.Error()
	} else {
		sf.SLogger().Error("need ptr")
		return ""
	}
}

func (sf *Controller) IsAdminManager(mid string) bool {
	return mid == "1" || mid == "585"
}

// 试题类型权限控制
func (sf *Controller) UserCategoryPermissionFilter(uid, category, childCate, subjectCate, keypoints string, cateType int, oldFilter any) (filter any) {
	filter = oldFilter
	if sf.IsAdminManager(uid) {
		return
	}
	var notExists = "no permission" // 不存在的值 用于搜索条件 使得查询为空
	if uid == "" {
		return
	}
	permissions := services.NewManualService().GetCategoryUsersPermissions(uid, string(rediskey.CategoryPermissionInterview))
	if len(permissions) == 0 {
		return "not set"
	}
	key := category
	if childCate != "" {
		key += "/" + childCate
	}

	// 处理 没选择分类的情况
	if cateType == 1 && category == "" {
		userCategory := make([]string, 0)
		for s, _ := range permissions {
			cateArr := strings.Split(s, "/")
			userCategory = append(userCategory, cateArr[0])
		}
		filter = bson.M{"$in": userCategory}
		return
	}

	// 处理 没选择分类的情况
	if cateType == 2 && subjectCate == "" {
		userCategory := make([]string, 0)
		for s, _ := range permissions {
			cateArr := strings.Split(s, "/")
			if category == "" {
				sub := common.GetMapKeys(permissions[s])
				if len(sub) > 0 {
					userCategory = append(userCategory, sub...)
				}
			} else if category == cateArr[0] {
				if childCate == "" || (len(cateArr) > 1 && childCate == cateArr[1]) {
					sub := common.GetMapKeys(permissions[s])
					if len(sub) > 0 {
						userCategory = append(userCategory, sub...)
					}
				}
			}
		}
		filter = bson.M{"$in": userCategory}
		return
	}

	// 处理没选择分类的情况
	if cateType == 3 && keypoints == "" {
		userCategory := make([]string, 0)
		for s, _ := range permissions {
			if category == "" {
				for _, subKeypoints := range permissions[s] {
					userCategory = append(userCategory, subKeypoints...)
				}
			} else if key == s {
				for sc, subKeypoints := range permissions[s] {
					if subjectCate != "" {
						if sc == subjectCate {
							userCategory = append(userCategory, subKeypoints...)
						} else {
							filter = notExists
							return
						}
					} else {
						userCategory = append(userCategory, subKeypoints...)
					}
				}
			}
		}
		filter = bson.M{"$in": userCategory}
		return
	}

	if cateType == 1 || cateType == 2 || cateType == 3 {
		if _, ok := permissions[key]; !ok {
			// 说明当前用户没有这个权限
			filter = notExists
			return
		}
	}

	if cateType == 2 || cateType == 3 {
		if _, ok := permissions[key][subjectCate]; !ok {
			// 说明当前用户没有这个权限
			filter = notExists
			return
		}
	}
	if cateType == 3 {
		if !common.InArrCommon(keypoints, permissions[key][subjectCate]) {
			// 说明当前用户没有这个权限
			filter = notExists
			return
		}
	}

	return
}

// 完善日期时间
func (sf *Controller) PerfectTimeFormat(s string, t int8) string {
	if strings.TrimSpace(s) != "" && len(strings.TrimSpace(s)) < 11 {
		if t == 1 {
			return fmt.Sprintf("%+v 00:00:00", s)
		} else {
			return fmt.Sprintf("%+v 59:59:59", s)
		}

	}
	return s
}
