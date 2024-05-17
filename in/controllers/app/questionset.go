package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"strconv"
)

type QuestionSet struct {
	controllers.Controller
}

func (sf *QuestionSet) HasReviewClass(c *gin.Context) {
	var resp []models.Review
	var err error
	var param struct {
		ClassId []string `json:"class_id"  binding:"required" `
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if len(param.ClassId) == 0 {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	//uid := c.GetHeader("X-XTJ-UID")

	resp = services.NewQuestionSet().ClassReview(param.ClassId)
	classIds := make(map[string]struct{})
	for _, review := range resp {
		classIds[review.Class.ClassID] = struct{}{}
	}

	sf.Success(common.GetMapKeys(classIds), c)
}

// 作业详情
func (sf *QuestionSet) ReviewWorkInfo(c *gin.Context) {
	var resp any
	var err error
	var id = c.Query("id") // review id
	if id == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 短ID
	if len(id) < 24 {
		review := new(models.Review)
		err = sf.DB().Collection("review").Where(bson.M{"short_id": id}).Take(review)
		if err != nil {
			if sf.MongoNoResult(err) {
				sf.Error(common.CodeInvalidParam, c, "测评不存在")
			} else {
				sf.Error(common.CodeInvalidParam, c, "服务器异常")
			}
			return
		}
		// 换成 长ID
		id = review.Id.Hex()
	}
	resp, err = services.NewQuestionSet().WorkInfo(id)

	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(resp, c)
}

func (sf *QuestionSet) ReviewInfo(c *gin.Context) {
	var resp any
	var err error
	var id = c.Query("id")                        // review id
	var studentId = c.Query("correct_student_id") // 1是老师
	var reAnswerStr = c.Query("re_answer")        // 重复创建一个review
	var preview = c.Query("preview")              //1 预览
	if id == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var reAnswer int8
	if reAnswerStr != "" {
		reAnswer = 1
	}
	uid := c.GetHeader("X-XTJ-UID")
	if studentId != "" {
		uid = studentId
	}
	// 短ID
	if len(id) < 24 {
		review := new(models.Review)
		err = sf.DB().Collection("review").Where(bson.M{"short_id": id}).Take(review)
		if err != nil {
			if sf.MongoNoResult(err) {
				sf.Error(common.CodeInvalidParam, c, "测评不存在")
			} else {
				sf.Error(common.CodeInvalidParam, c, "服务器异常")
			}
			return
		}
		// 换成 长ID
		id = review.Id.Hex()
	}
	if studentId != "" {
		resp, err = services.NewQuestionSet().LogInfo(uid, id)
	} else {
		resp, err = services.NewQuestionSet().MakeLog(uid, id, reAnswer, preview)
	}
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(resp, c)
}

func (sf *QuestionSet) ReviewList(c *gin.Context) {
	var err error
	var uid = c.Query("student_id")
	var classId = c.Query("class_id")
	var teacher = c.Query("teacher_correct") // 1是老师
	var page_index = c.Query("page_index")
	var page_size = c.Query("page_size")
	if classId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	PageIndex, _ := strconv.ParseInt(page_index, 10, 64)
	PageSize, _ := strconv.ParseInt(page_size, 10, 64)
	offset, limit := sf.PageLimit(PageIndex, PageSize)
	list, total, err := services.NewQuestionSet().ClassReviewList(uid, classId, teacher, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]any{
		"list":  list,
		"total": total,
	}, c)
}

func (sf *QuestionSet) ReviewLogList(c *gin.Context) {
	var err error
	var page_index = c.Query("page_index")
	var page_size = c.Query("page_size")
	uid := c.GetHeader("X-XTJ-UID")
	PageIndex, _ := strconv.ParseInt(page_index, 10, 64)
	PageSize, _ := strconv.ParseInt(page_size, 10, 64)
	offset, limit := sf.PageLimit(PageIndex, PageSize)
	list, total, err := services.NewQuestionSet().ReviewLogList(uid, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]any{
		"list":  list,
		"total": total,
	}, c)
}

func (sf *QuestionSet) GetCorrectComment(c *gin.Context) {
	list := services.NewQuestionSet().GetCorrectComment()
	sf.Success(list, c)
}

// SaveAnswerComment 保存点评
func (sf *QuestionSet) SaveAnswerComment(c *gin.Context) {
	var param struct {
		ReviewLogId string                       `json:"review_log_id"  binding:"required" `   //测评记录id
		AnswerLogId string                       `json:"g_answer_log_id"  binding:"required" ` //回答记录id
		Comment     models.TeacherCorrectComment `json:"comment"  binding:"required" `         //评论内容
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.ReviewLogId == "" || param.AnswerLogId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	list := services.NewQuestionSet().GetCorrectComment()
	mlist := make(map[string][]string)
	for _, item := range list {
		mlist[item.Key] = item.Values
	}

	if !common.InArrCommon[string](param.Comment.Content, mlist["content"]) {
		sf.Error(common.InvalidId, c, "点评中的内容不符合要求")
		return
	}

	if !common.InArrCommon[string](param.Comment.Speed, mlist["speed"]) {
		sf.Error(common.InvalidId, c, "点评中的语速不符合要求")
		return
	}
	if !common.InArrCommon[string](param.Comment.Interaction, mlist["interaction"]) {
		sf.Error(common.InvalidId, c, "点评中的互动不符合要求")
		return
	}
	if !common.InArrCommon[string](param.Comment.Confident, mlist["confident"]) {
		sf.Error(common.InvalidId, c, "点评中的自信不符合要求")
		return
	}
	commentKey := param.Comment.Content + "-" + param.Comment.Speed + "-" + param.Comment.Interaction + "-" + param.Comment.Confident
	grade := common.GetGrade(commentKey)
	param.Comment.Grade = grade

	if param.Comment.GAnswer.VoiceUrl != "" {
		param.Comment.GAnswer.VoiceLength = sf.TransitionFloat64(param.Comment.GAnswer.VoiceLength, -1)
	}

	services.NewQuestionSet().Correct(uid, param.ReviewLogId, param.AnswerLogId, param.Comment)

	sf.Success("success", c)
}
