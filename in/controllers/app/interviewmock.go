package app

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"strconv"
	"time"
)

type InterviewMockController struct {
	controllers.Controller
}

type ImReqParams struct {
	PageIndex     int64  `json:"page_index"`
	PageSize      int64  `json:"page_size"`
	ExamStartTime string `json:"exam_start_time"`
	ExamEndTime   string `json:"exam_end_time"`
	MockTime      string `json:"mock_time"`
}

type InterviewMockInfo struct {
	ID                string                 `json:"id" bson:"_id,omitempty"`
	Title             string                 `json:"title"`
	Desc              string                 `json:"desc"`
	CurriculaID       string                 `json:"curricula_id" bson:"curricula_id"`
	ExamStartTime     string                 `json:"exam_start_time" bson:"exam_start_time"`
	ExamEndTime       string                 `json:"exam_end_time" bson:"exam_end_time"`
	MockTime          string                 `json:"mock_time" bson:"mock_time"`
	OneByOneTimeType  int                    `json:"onebyone_time_type" bson:"onebyone_time_type"`
	SimulateTimeType  int                    `json:"simulate_time_type" bson:"simulate_time_type"`
	Creator           string                 `json:"creator" bson:"creator"`
	QuestionnaireList []models.Questionnaire `json:"questionnaire_list" bson:"questionnaire_list"`
}

type ListResp struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	CurriculaTitle   string `json:"curricula_title"`
	ExamStartTime    string `json:"exam_start_time" bson:"exam_start_time"`
	ExamEndTime      string `json:"exam_end_time" bson:"exam_end_time"`
	MockTime         string `json:"mock_time" bson:"mock_time"`
	CurriculaID      string `json:"curricula_id"`
	OneByOneTimeType int    `json:"onebyone_time_type" bson:"onebyone_time_type"`
	SimulateTimeType int    `json:"simulate_time_type" bson:"simulate_time_type"`
	IsApply          bool   `json:"is_apply"`
	Status           int    `json:"status"`
}
type ListResponse struct {
	Total int64      `json:"total"`
	List  []ListResp `json:"list"`
}

// CreatePermission 用户是否可以创建模考
func (i *InterviewMockController) CreatePermission(c *gin.Context) {
	id := c.DefaultQuery("id", "")
	userId := c.GetHeader("X-XTJ-UID")
	// 检测权限是否是考试的管理员
	flag := models.CurriculaModel.CheckCurriculaAdmin(userId, id)
	i.Success(flag, c)
}

// List 模考列表 先区分角色,考试的管理员 只能看到自己的管理的模考
func (i *InterviewMockController) List(c *gin.Context) {
	var err error
	index, _ := c.GetQuery("page_index")
	size, _ := c.GetQuery("page_size")
	indexInt, _ := strconv.ParseInt(index, 10, 64)
	sizeInt, _ := strconv.ParseInt(size, 10, 64)
	offset, limit := i.PageLimit(indexInt, sizeInt)
	var result ListResponse
	// 获取用户属于的考试 用户管理的考试下的模考 和 所属班级的所属的考试下的模考 还有已经报名的有可能不属于前两种情况的模考
	userID := c.GetHeader("X-XTJ-UID")
	var list []models.InterviewMock
	applyMap := make(map[string]bool)
	currMap := make(map[string]models.Curricula)
	//  获取自己的角色,考试管理员只能看到自己管理的模考
	userControlCuIDs := models.CurriculaModel.GetUserControlCurrID(userID)
	if len(userControlCuIDs) > 0 {
		// 考试管理员
		f := bson.M{"is_deleted": 0, "curricula_id": bson.M{"$in": userControlCuIDs}}
		list, err = models.InterviewMockModel.GetList(f)
		currMap = models.CurriculaModel.GetCurriculaMap(userControlCuIDs)
	} else {
		// 普通用户
		currIDs, applyMockID := models.InterviewMockModel.GetUserCurrIDAndApplyMockID(userID)
		list, applyMap, err = models.InterviewMockModel.GetUserApplyList(currIDs, applyMockID)
		ids := make([]string, 0, len(list))
		for _, v := range list {
			ids = append(ids, v.CurriculaID)
		}
		currMap = models.CurriculaModel.GetCurriculaMap(ids)
	}
	if err != nil {
		i.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	result.Total = int64(len(list))
	result.List = make([]ListResp, 0, len(list))
	if int64(len(list)) < offset {
		i.Success(result, c)
		return
	}
	last := offset + limit
	if last >= int64(len(list)) {
		last = int64(len(list))
	}
	list = list[offset:last]
	var resp []ListResp
	for _, item := range list {
		tmp := ListResp{
			ID:               item.Id.Hex(),
			Title:            item.Title,
			ExamStartTime:    item.ExamStartTime,
			ExamEndTime:      item.ExamEndTime,
			MockTime:         item.MockTime,
			CurriculaID:      item.CurriculaID,
			OneByOneTimeType: item.OneByOneTimeType,
			SimulateTimeType: item.SimulateTimeType,
		}
		if v, ok := currMap[item.CurriculaID]; ok {
			tmp.CurriculaTitle = v.CurriculaTitle
		}
		if _, ok := applyMap[tmp.ID]; ok {
			tmp.IsApply = true
		}
		tmp.Status = item.Status()
		resp = append(resp, tmp)
	}
	result.List = resp
	i.Success(result, c)
}
func (i *InterviewMockController) convert(param InterviewMockInfo) models.InterviewMock {
	res := models.InterviewMock{
		Title:             param.Title,
		Desc:              param.Desc,
		CurriculaID:       param.CurriculaID,
		ExamStartTime:     param.ExamStartTime,
		ExamEndTime:       param.ExamEndTime,
		MockTime:          param.MockTime,
		OneByOneTimeType:  param.OneByOneTimeType,
		SimulateTimeType:  param.SimulateTimeType,
		Creator:           param.Creator,
		QuestionnaireList: param.QuestionnaireList,
	}
	if param.ID != "" {
		res.Id = i.ObjectID(param.ID)
	}
	return res
}

// Create 创建模考
func (i *InterviewMockController) Create(c *gin.Context) {
	var err error
	var param InterviewMockInfo
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c)
		return
	}
	userId := c.GetHeader("X-XTJ-UID")
	// 是否有权限操作
	if !models.CurriculaModel.CheckCurriculaAdmin(userId, param.CurriculaID) {
		i.Error(common.PermissionDenied, c)
		return
	}
	param.Creator = userId
	err = models.InterviewMockModel.Create(i.convert(param))
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeServerBusy, c)
		return
	}
	i.Success(nil, c)
}

// Update 更新模考
func (i *InterviewMockController) Update(c *gin.Context) {
	var err error
	var param InterviewMockInfo
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c)
		return
	}
	userId := c.GetHeader("X-XTJ-UID")
	// 是否有权限操作
	if !models.CurriculaModel.CheckCurriculaAdmin(userId, param.CurriculaID) {
		i.Error(common.PermissionDenied, c)
		return
	}
	param.Creator = userId
	err = models.InterviewMockModel.Update(i.convert(param))
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeServerBusy, c)
		return
	}
	i.Success(nil, c)
}

// Delete 删除模考
func (i *InterviewMockController) Delete(c *gin.Context) {
	var err error
	var param InterviewMockInfo
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c)
		return
	}
	if param.ID == "" {
		i.Error(common.CodeInvalidParam, c, "id is nil")
		return
	}
	info, err := models.InterviewMockModel.GetInfo(i.ObjectID(param.ID))
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	userId := c.GetHeader("X-XTJ-UID")
	// 是否有权限操作
	if !models.CurriculaModel.CheckCurriculaAdmin(userId, info.CurriculaID) {
		i.Error(common.PermissionDenied, c)
		return
	}
	err = models.InterviewMockModel.Delete(i.ObjectID(param.ID))
	fmt.Println(err)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeServerBusy, c)
		return
	}
	i.Success(nil, c)
}

// Info 模考详情
func (i *InterviewMockController) Info(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	info, err := models.InterviewMockModel.GetInfo(i.ObjectID(id))
	if errors.Is(err, mongo.ErrNoDocuments) {
		i.Error(common.Closed, c, "模考不存在")
		return
	}
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeServerBusy, c, err.Error())
		return
	}

	i.Success(info, c)
}

func (i *InterviewMockController) UnApplyList(c *gin.Context) {
	userID := c.GetHeader("X-XTJ-UID")
	if userID == "" {
		i.Success(nil, c, "un login")
		return
	}
	currIDs, applyMockIDs := models.InterviewMockModel.GetUserCurrIDAndApplyMockID(userID)
	f := bson.M{
		"is_deleted":      0,
		"curricula_id":    bson.M{"$in": currIDs},
		"exam_start_time": bson.M{"$lte": time.Now().Format("2006-01-02 15:04")},
		"exam_end_time":   bson.M{"$gte": time.Now().Format("2006-01-02 15:04")},
	}
	list, err := models.InterviewMockModel.GetList(f)
	if err != nil {
		i.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	applyIDsMap := make(map[string]bool)
	for _, id := range applyMockIDs {
		applyIDsMap[id.Hex()] = true
	}
	// 只保留未报名的
	resp := make([]ListResp, 0, len(list))
	for _, item := range list {
		// 已报名的过滤
		if _, ok := applyIDsMap[item.Id.Hex()]; ok {
			continue
		}
		tmp := ListResp{
			ID:            item.Id.Hex(),
			Title:         item.Title,
			ExamStartTime: item.ExamStartTime,
			ExamEndTime:   item.ExamEndTime,
			MockTime:      item.MockTime,
			CurriculaID:   item.CurriculaID,
			IsApply:       false,
			Status:        item.Status(),
		}
		resp = append(resp, tmp)
	}
	i.Success(resp, c)
}
