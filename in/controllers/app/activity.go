package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/helper"
	"interview/models"
	"interview/services"
	"regexp"
	"time"
)

type Activity struct {
	controllers.Controller
}

// UserInviteList 用户邀请过的用户列表
func (sf *Activity) UserInviteList(c *gin.Context) {
	var err error
	uid := c.GetHeader("X-XTJ-UID")
	filter := bson.M{"user_id": uid}
	totalCount, _ := sf.DB().Collection("invite_user").Where(filter).Count()
	var iu = []models.InviteUser{}
	err = sf.DB().Collection("invite_user").Where(filter).Sort("-created_time").Find(&iu)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	list := []models.InviteUser{}
	ids := []string{}
	for _, i := range iu {
		ids = append(ids, i.InvitedUserId)
	}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo(ids, appCode, 1)
	for _, i := range iu {
		i.InvitedUserName = userMap[i.InvitedUserId].Nickname
		i.InvitedUserAvatar = userMap[i.InvitedUserId].Avatar
		list = append(list, i)
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

// SaveInviteInfo 邀请好友后保存邀请用户记录和GPT使用次数
func (sf *Activity) SaveInviteInfo(c *gin.Context) {
	inviterID := c.DefaultQuery("uid", "")
	if inviterID == "" {
		sf.Error(common.CodeServerBusy, c, "邀请链接有误，请核实")
		return
	}
	invitedUID := c.GetHeader("X-XTJ-UID")
	if inviterID == invitedUID {
		sf.Error(common.CodeServerBusy, c, "邀请人和被邀请人不能是同一人！")
		return
	}
	//判断用户是否已被同一人邀请
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	filter := bson.M{"user_id": inviterID, "invited_user_id": invitedUID}
	totalCount, _ := sf.DB().Collection("invite_user").Where(filter).Count()
	if totalCount == 0 {
		var iu = models.InviteUser{UserId: inviterID, InvitedUserId: invitedUID}
		_, err := sf.DB().Collection("invite_user").Create(&iu)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
			return
		}
	} else {
		sf.Error(common.CodeServerBusy, c, "你已被ta邀请过啦！")
		return
	}

	// 邀请一人得一次提问次数
	// 查询GPT次数
	countInfo, err := sf.GetGPTCanUseCount(inviterID)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
		return
	}
	// 修改GPT次数
	if countInfo.TotalCount == 0 {
		countInfo.TotalCount += 2
		countInfo.AvailableCount += 2
	} else {
		countInfo.TotalCount += 1
		countInfo.AvailableCount += 1
	}
	countInfo.TotalInviteCount += 1
	err = sf.SetGPTCanUseCount(countInfo)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
		return
	}

	sf.Success(nil, c)
}

func (sf *Activity) QueryGPTUseCount(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	countInfo, err := sf.GetGPTCanUseCount(uid)
	if err != nil {
		sf.Error(common.CodeServerBusy, c)
	} else {
		sf.Success(map[string]interface{}{"count_info": countInfo}, c)
	}
}

// SubmitCustomQuestion 学员提问
func (sf *Activity) SubmitCustomQuestion(c *gin.Context) {
	var param struct {
		Name              string   `json:"name" binding:"required"` // 试题名称
		AnswerType        int8     `json:"answer_type"`
		ExamCategory      string   `json:"exam_category"`       // 问题分类，如事业单位，教招面试，教资面试
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   //题分类
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	// 判断学员是否还有GPT使用次数
	countInfo, err := sf.GetGPTCanUseCount(uid)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
		return
	} else {
		if countInfo.AvailableCount <= 0 {
			sf.Error(common.CodeServerBusy, c, "已经没有可用次数啦，快去邀请用户获得次数吧~")
			return
		}
	}
	// GPT可用次数-1
	// 如果total_count为1，代表此时redis还没有数据
	countInfo.AvailableCount -= 1
	err = sf.SetGPTCanUseCount(countInfo)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
		return
	}

	var question models.GCustomQuestion
	// 新增试题
	question.UserId = uid
	question.Name = param.Name
	question.AnswerType = param.AnswerType
	question.ExamCategory = param.ExamCategory
	question.ExamChildCategory = param.ExamChildCategory
	question.QuestionCategory = param.QuestionCategory
	resp, err := sf.DB().Collection("g_custom_questions").Create(&question)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	id := ""
	if _id, ok := resp.InsertedID.(primitive.ObjectID); ok {
		id = _id.Hex()
	}
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	rdb.Do("RPUSH", rediskey.GPTCustomWaiting, id)
	sf.Success(map[string]string{"content": id}, c)
}

// CustomQuestionList 查看自主提问试题
func (sf *Activity) CustomQuestionList(c *gin.Context) {
	var param struct {
		ExamCategory string `json:"exam_category"`
		AnswerType   int8   `json:"answer_type"`
		PageIndex    int64  `json:"page_index"`
		PageSize     int64  `json:"page_size"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	appCode := c.GetString("APP-CODE")
	filter := bson.M{"user_id": uid, "answer_type": param.AnswerType}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	totalCount, _ := sf.DB().Collection("g_custom_questions").Where(filter).Count()
	var questions = []models.GCustomQuestion{}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_custom_questions").Where(filter).Sort("-created_time").Skip(offset).Limit(limit).Find(&questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	list := []models.GCustomQuestion{}
	ids := []string{}
	for _, q := range questions {
		ids = append(ids, q.UserId)
	}
	userMap := new(models.InterviewGPT).GetUsersInfo(ids, appCode, 1)
	for _, q := range questions {
		q.UserName = userMap[q.UserId].Nickname
		q.UserAvatar = userMap[q.UserId].Avatar
		list = append(list, q)
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)

}

// CustomQuestionInfo GetInterviewQuestion 查看试题详情
func (sf *Activity) CustomQuestionInfo(c *gin.Context) {
	QuestionID := c.Query("question_id")
	if QuestionID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var question models.GCustomQuestion
	filter := bson.M{"_id": sf.ObjectID(QuestionID)}
	err := sf.DB().Collection("g_custom_questions").Where(filter).Take(&question)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "试题不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	// 添加昵称和头像
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo([]string{question.UserId}, appCode, 1)
	question.UserName = userMap[question.UserId].Nickname
	question.UserAvatar = userMap[question.UserId].Avatar
	sf.Success(question, c)
}

// GetGPTCanUseCount 从mongo查询GPT可用次数
func (sf *Activity) GetGPTCanUseCount(uid string) (models.GPTCountInfo, error) {
	var countInfo models.GPTCountInfo
	err := sf.DB().Collection("gpt_count_info").Where(bson.M{"user_id": uid}).Take(&countInfo)
	if err != nil {
		if sf.MongoNoResult(err) {
			var TotalCount = 10
			var AvailableCount = 10
			if helper.RedisSIsMember("interview:temporary_uids", uid) {
				TotalCount = 1000
				AvailableCount = 1000
			}
			countInfo.UserID = uid
			countInfo.TotalCount = TotalCount
			countInfo.AvailableCount = AvailableCount
			countInfo.TotalInviteCount = 0
			countInfo.SendCount = 0
			countInfo.BaipiaoCount = 0
			_, err = sf.DB().Collection("gpt_count_info").Create(&countInfo)
			if err != nil {
				sf.SLogger().Error(err)
				return countInfo, err
			} else {
				return countInfo, nil
			}
		} else {
			sf.SLogger().Error(err)
			return countInfo, err
		}
	} else {
		if helper.RedisSIsMember("interview:temporary_uids", uid) {
			countInfo.TotalCount = 1000
			countInfo.AvailableCount = 1000
		}
		return countInfo, nil
	}
}

// SetGPTCanUseCount 从mongo修改GPT可用次数
func (sf *Activity) SetGPTCanUseCount(info models.GPTCountInfo) error {
	userID := info.UserID
	if helper.RedisSIsMember("interview:temporary_uids", userID) {
		return nil
	}
	err := sf.DB().Collection("gpt_count_info").Where(bson.M{"user_id": userID}).Save(&info)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (sf *Activity) NeedBaipiao(c *gin.Context) {
	var param struct {
		Reason string `json:"reason"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.Reason == "" {
		sf.Error(common.CodeInvalidParam, c, "请输入原因~")
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var t models.BaiPiao
	t.UserID = uid
	today := time.Now().Format("2006-01-02")
	filter := bson.M{"user_id": uid, "created_time": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(today)}}}
	err = sf.DB().Collection("baipiao_logs").Where(filter).Take(&t)
	if err == nil {
		sf.Success("今日已白嫖过了，请明日再来吧~", c)
		return
	}
	t.Reason = param.Reason
	t.Count = 10
	_, err = sf.DB().Collection("baipiao_logs").Create(&t)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 增加次数
	countInfo, err := sf.GetGPTCanUseCount(uid)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	countInfo.BaipiaoCount += 10
	countInfo.BaipiaoTimeCount += 1
	countInfo.TotalCount += 10
	countInfo.AvailableCount += 10
	err = sf.SetGPTCanUseCount(countInfo)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success("成功白嫖，次数+10~", c)
}

// IsTodayHaveBaipiao 今天是否已经白嫖过了
func (sf *Activity) IsTodayHaveBaipiao(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	var t models.BaiPiao
	t.UserID = uid
	today := time.Now().Format("2006-01-02")
	filter := bson.M{"user_id": uid, "created_time": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(today)}}}
	err := sf.DB().Collection("baipiao_logs").Where(filter).Take(&t)
	if err == nil {
		sf.Success(map[string]bool{"is_today_have_baipioa": true}, c)
	} else {
		if sf.MongoNoResult(err) {
			sf.Success(map[string]bool{"is_today_have_baipioa": false}, c)
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
}

// NeedBuyCount 购买次数
func (sf *Activity) NeedBuyCount(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	countInfo, err := sf.GetGPTCanUseCount(uid)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
		return
	}
	_, err = sf.DB().Collection("gpt_count_info").Where(bson.M{"user_id": uid}).Update(bson.M{"buy_time_count": countInfo.BuyTimeCount + 1})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

func (sf *Activity) TransferID(c *gin.Context) {
	var param struct {
		QuestionID string `json:"question_id"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	id, err := new(InterviewGPT).TransferIDLength(param.QuestionID)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	} else {
		sf.Success(map[string]string{"id": id}, c)
	}
}

func (sf *Activity) Area(c *gin.Context) {
	sf.Success(new(services.ServicesBase).AreaJson(), c)
}
