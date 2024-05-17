package manager

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"interview/router/request"
	"interview/services"
	"regexp"
	"strconv"
)

type MockExam struct {
	controllers.Controller
}

func (sf *MockExam) List(c *gin.Context) {
	var param struct {
		Keywords     string `json:"keywords"`
		ExamCategory string `json:"exam_category"`
		PageIndex    int64  `json:"page_index"`
		PageSize     int64  `json:"page_size"`
		Status       int32  `json:"status"` // 试题状态
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"exam_type": bson.M{"$ne": 1}}

	if param.Status != 0 {
		filter["status"] = param.Status
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
	resultInfo, total, err := services.NewMockExamService().List(filter, offset, limit)

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

func (sf *MockExam) Edit(c *gin.Context) {
	var param request.MockexamEditRequest
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	id, err := services.NewMockExamService().Edit(param)
	if err != nil {
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(id, c)
}

func (sf *MockExam) SlotTime(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	examTimeStr := c.Query("exam_time") // 秒
	restTimeStr := c.Query("rest_time") // 秒

	if startTimeStr == "" || endTimeStr == "" || examTimeStr == "" || restTimeStr == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	startTime, _ := common.LocalTimeFromDateString(startTimeStr)
	endTime, _ := common.LocalTimeFromDateString(endTimeStr)

	examTime, _ := strconv.Atoi(examTimeStr)
	restTime, _ := strconv.Atoi(restTimeStr)
	slotArr := services.NewMockExamService().SlotTime(startTime, endTime, examTime, restTime)

	sf.Success(slotArr, c)
}

func (sf *MockExam) GetTeacher(c *gin.Context) {
	uid := c.Query("user_id")
	userMap := new(models.InterviewGPT).GetUsersInfo([]string{uid}, "402", 1)
	var user models.GUser
	if _, ok := userMap[uid]; ok {
		user = userMap[uid]
	}
	sf.Success(user, c)
}
