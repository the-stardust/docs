package app

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/helper"
	"interview/router/request"
	"interview/services"
	"regexp"
	"strconv"
	"time"
)

type MockExam struct {
	controllers.Controller
}

func (sf *MockExam) BaseInfo(c *gin.Context) {
	var err error
	examId := c.Query("exam_id")
	uid := c.GetHeader("X-XTJ-UID")

	exam, err := services.NewMockExamService().Info(uid, examId, "", "0")
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(exam, c)
}
func (sf *MockExam) Info(c *gin.Context) {
	var err error
	examId := c.Query("exam_id")
	uid := c.GetHeader("X-XTJ-UID")
	roomId := c.Query("room_id")

	exam, err := services.NewMockExamService().Info(uid, examId, roomId, "1")
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(exam, c)
}

// TeacherMarkContent
func (sf *MockExam) TeacherMarkContent(c *gin.Context) {
	sf.Success(services.NewMockExamService().GetMarkComment(), c)
}

// TeacherMark
func (sf *MockExam) TeacherMark(c *gin.Context) {
	var param request.TeacherMarkRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	uid := c.GetHeader("X-XTJ-UID")

	err = services.NewMockExamService().TeacherMark(uid, param)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// RoomList
func (sf *MockExam) RoomList(c *gin.Context) {
	var err error
	//examId := c.Query("exam_id")
	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")
	if pageIndexStr == "" {
		pageIndexStr = "1"
	}
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageIndex, _ := strconv.Atoi(pageIndexStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))

	uid := c.GetHeader("X-XTJ-UID")
	//err = sf.UserSets(uid)
	//if err != nil {
	//	sf.Success(map[string]any{
	//		"list":  make([]int, 0),
	//		"total": 0,
	//	}, c)
	//	return
	//}

	list, total, err := services.NewMockExamService().LogList(uid, offset, limit)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]any{
		"list":  list,
		"total": total,
	}, c)
}

// RoomStatus
func (sf *MockExam) RoomStatus(c *gin.Context) {
	roomId := c.Query("room_id")
	//uid := c.GetHeader("X-XTJ-UID")
	cacheKey := string(rediskey.MockExamStatus) + roomId
	status, _ := helper.RedisHGetAll(cacheKey)
	var userControl any
	if _, ok := status["control"]; ok {
		controlStr, _ := helper.RedisHGet(cacheKey, "control")
		var control = make(map[string]map[string]int)
		if controlStr != "" {
			json.Unmarshal([]byte(controlStr), &control)
		}
		delete(status, "control")
		//if _, ok := control[uid]; ok {
		//	userControl = control[uid]
		//} else {
		userControl = control
		//}
	}
	sf.Success(map[string]any{
		"status":  status,
		"control": userControl,
	}, c)
}

// SetRoomStatus
func (sf *MockExam) SetRoomStatus(c *gin.Context) {
	var param request.RoomStatusRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	roomId := param.RoomId
	action := param.Action // cur_uid/cur_qid/step/close/firstteacher
	val := param.Val
	//uid := c.GetHeader("X-XTJ-UID")

	cacheKey := string(rediskey.MockExamStatus) + roomId
	if action == "firstteacher" {
		teacherId, _ := helper.RedisHGet(cacheKey, action)
		if teacherId == "" {
			helper.RedisHSet(cacheKey, action, []byte(val))
			helper.RedisEXPIRE(cacheKey, 14400)
		}
	} else {
		helper.RedisHSet(cacheKey, action, []byte(val))
		helper.RedisEXPIRE(cacheKey, 14400)
	}
	if action == "start" {
		sf.DB().Collection("mock_exam_slot_room").Where(bson.M{"_id": sf.ObjectID(roomId)}).Update(bson.M{"status": 2})
	}
	if action == "close" {
		sf.DB().Collection("mock_exam_slot_room").Where(bson.M{"_id": sf.ObjectID(roomId)}).Update(bson.M{"status": 1})
	}
	if action == "cur_qid" {
		_ = services.NewMockExamService().SetRoomUseQuestion(roomId, val)
	}

	sf.Success("success", c)
}

func (sf *MockExam) LogDetail(c *gin.Context) {
	var err error
	roomId := c.Query("room_id")
	uid := c.Query("user_id")
	if uid == "" {
		uid = c.GetHeader("X-XTJ-UID")
	}
	logList, err := services.NewMockExamService().LogDetail(uid, roomId)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(logList, c)
}

func (sf *MockExam) UserSets(uid string) error {
	// 是否是指定的用户
	_switch, _ := helper.RedisGet("interview:mock_exam_switch")
	if _switch != "" {
		if uid == "" || !helper.RedisSIsMember("interview:mock_exam_user_sets", uid) {
			return errors.New("not in")
		}
	}
	return nil
}

func (sf *MockExam) List(c *gin.Context) {
	var param struct {
		Keywords     string `json:"keywords"`
		ExamCategory string `json:"exam_category"`
		PageIndex    int64  `json:"page_index"`
		PageSize     int64  `json:"page_size"`
		ExamType     int8   `json:"exam_type"` // 0 普通考试，1 自主练习
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	err = sf.UserSets(uid)
	if err != nil {
		sf.Success(map[string]any{
			"list":  make([]int, 0),
			"total": 0,
		}, c)
		return
	}

	filter := bson.M{}
	filter["status"] = 1
	if param.ExamType != 0 {
		filter["exam_type"] = param.ExamType
	} else {
		filter["exam_type"] = bson.M{"$ne": 1} // 不等于自主练习
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}

	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"title": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"sub_title": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}}, bson.M{"_id": param.Keywords},
			//bson.M{"question_source": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}

	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	resultInfo, total, err := services.NewMockExamService().List2(filter, uid, offset, limit)

	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]any{
		"list":  resultInfo,
		"total": total,
	}, c)

}

// 报名
func (sf *MockExam) SignUp(c *gin.Context) {
	var param request.SignUpRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	err = sf.UserSets(uid)
	if err != nil {
		sf.Success(map[string]any{
			"list":  make([]int, 0),
			"total": 0,
		}, c)
		return
	}

	log, err := services.NewMockExamService().SignUp(uid, param, 3)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(log.RoomId, c)
}

// Cancel
func (sf *MockExam) Cancel(c *gin.Context) {
	mockExamId := c.Query("exam_id")
	uid := c.Query("uid")
	if uid == "" {
		uid = c.GetHeader("X-XTJ-UID")
	}
	if mockExamId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	err := services.NewMockExamService().Cancel(uid, mockExamId)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// 签到
func (sf *MockExam) SignIn(c *gin.Context) {
	mockExamId := c.Query("exam_id")
	uid := c.GetHeader("X-XTJ-UID")
	if mockExamId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	err := services.NewMockExamService().SignIn(uid, mockExamId)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// 订阅
func (sf *MockExam) Subscribe(c *gin.Context) {
	var param struct {
		ExamId string `json:"exam_id" form:"exam_id"`
		Oid    string `json:"oid" form:"oid"` // user openid
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")

	err = services.NewMockExamService().Subscribe(uid, param.Oid, param.ExamId)
	if err != nil {
		sf.Logger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// End
func (sf *MockExam) End(c *gin.Context) {
	roomId := c.Query("room_id")
	uid := c.GetHeader("X-XTJ-UID")
	if roomId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	err := services.NewMockExamService().End(uid, roomId)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// 房间token
func (sf *MockExam) RoomToken(c *gin.Context) {
	var param request.ZegoTokenRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.RoomId == "" || param.Uid == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	token, err := services.NewZego().GetPrivilegeToken(param.RoomId, param.Uid)
	if err != nil {
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(token, c)
}

// 老师控制
func (sf *MockExam) TeacherControl(c *gin.Context) {
	var param request.TeacherControlRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	if param.RoomId == "" || len(param.Action) == 0 {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	err = services.NewMockExamService().TeacherControl(uid, param)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

// 获取老师控制
//func (sf *MockExam) GetTeacherControl(c *gin.Context) {
//	roomId := c.Query("room_id")
//	uid := c.GetHeader("X-XTJ-UID")
//	sf.Success(services.NewMockExamService().GetTeacherControl(uid, roomId), c)
//}

// IsTeacher
func (sf *MockExam) IsTeacher(c *gin.Context) {
	uid := c.Query("user_id")
	examId := c.Query("exam_id")
	sf.Success(services.NewMockExamService().IsTeacher(uid, examId), c)
}

// 是否自动录制
func (sf *MockExam) AutoRecord(c *gin.Context) {
	sf.Success(common.InArrCommon[string](time.Now().Format("2006-01-02"), []string{"2024-04-23", "2024-04-24", "2024-04-25", "2024-04-26"}), c)
}

// ZegoCallback 云端录制的回调
// doc: https://doc-zh.zego.im/article/12324#4_1
func (sf *MockExam) ZegoCallback(c *gin.Context) {
	var param struct {
		AppId     int64          `json:"app_id" form:"app_id"`
		TaskId    string         `json:"task_id" form:"task_id"`
		RoomId    string         `json:"room_id" form:"room_id"`
		Message   string         `json:"message" form:"message"`
		EventType int            `json:"event_type" form:"event_type"`
		Nonce     string         `json:"nonce" form:"nonce"`
		Timestamp string         `json:"timestamp" form:"timestamp"`
		Signature string         `json:"signature" form:"signature"`
		Detail    map[string]any `json:"detail" form:"detail"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	sf.SLogger().Info("ZegoCallback:", param)
	if !services.NewZego().CheckSign(param.Signature, param.Nonce, param.Timestamp) {
		sf.SLogger().Error("zego CheckSign fail: ", "Signature:"+param.Signature, " Nonce:"+param.Nonce+" Timestamp:", param.Timestamp)
		sf.Error(common.PermissionDenied, c)
		return
	}

	if param.EventType == 1 {
		if statusInt, ok := param.Detail["upload_status"].(float64); ok && statusInt == 1 {
			if fi, ok := param.Detail["file_info"]; ok {
				if ffi, ok := fi.([]any); ok {
					if fffi, ok := ffi[0].(map[string]any); ok {
						services.NewMockExamService().SetExamRoomVideoUrl(param.RoomId, param.TaskId, fffi["file_url"].(string))
					}
				}
			}
		}
	}

	sf.Success("success", c)
}

// StartVideo
func (sf *MockExam) StartVideo(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	roomId := c.Query("room_id")
	//uid = "cfvc3d432ur4tf5j2umg"
	taskId, err := services.NewZego().StartVideoRecord(roomId, uid)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	err = services.NewMockExamService().SetExamRoomVideoTaskId(roomId, taskId)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	sf.Success(taskId, c)
}

// EndVideo
func (sf *MockExam) EndVideo(c *gin.Context) {
	roomId := c.Query("room_id")
	room := services.NewMockExamService().GetExamRoom(roomId)
	err := services.NewZego().EndVideoRecord(room.VideoTaskId)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	sf.Success("success", c)
}

func (sf *MockExam) RoomMemberChange(c *gin.Context) {
	var param request.RoomMemberChangeRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	err = services.NewMockExamService().RoomMemberChange(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	sf.Success("success", c)

}

// RoomConf
func (sf *MockExam) RoomConf(c *gin.Context) {
	appId, _ := strconv.ParseUint(global.CONFIG.Zego.AppId, 10, 64)
	yearMonth, _ := strconv.ParseUint(time.Now().Format("200601"), 0, 64)
	appId += yearMonth
	sign := global.CONFIG.Zego.AppSign

	sf.Success(map[string]any{
		"app_id":   appId,                                               // +年月
		"app_sign": sign[len(sign)-4:] + sign[4:len(sign)-4] + sign[:4], // 前后四位字符串互换
		"room_url": global.CONFIG.Zego.RoomUrl,
	}, c)
}

// 自主练习房间创建
func (sf *MockExam) RoomCreate(c *gin.Context) {
	var param request.MockExamCreateRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if len([]rune(param.Title)) > 64 {
		sf.Error(common.CodeInvalidParam, c, "标题最大64位")
		return
	}
	if len([]rune(param.Password)) > 64 {
		sf.Error(common.CodeInvalidParam, c, "密码最大64位")
		return
	}
	param.Uid = c.GetHeader("X-XTJ-UID")

	exam, room, err := services.NewMockExamService().Create(param)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]any{
		"exam": exam,
		"room": room,
	}, c)
}

// 中途加入
func (sf *MockExam) VideoStartMidwayJoin(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	roomId := c.Query("room_id")
	err := services.NewMockExamService().VideoStartMidwayJoin(roomId, uid)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}
