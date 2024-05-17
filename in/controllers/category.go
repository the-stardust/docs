package controllers

import (
	"encoding/json"
	"interview/common"
	"interview/models"
	"interview/models/appresp"
	"interview/models/managerresp"
	"interview/services"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
)

type Category struct {
	Controller
}

// 考试分类
func (sf *Category) ExamCategory(c *gin.Context) {
	var err error
	var examCategorys = make([]models.ExamCategory, 0)
	err = sf.DB().Collection("exam_category").Sort("-sort_number").Find(&examCategorys)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	openCategoryPermission := c.Query("open_category_permission")
	// 走权限
	if openCategoryPermission == "1" {
		uid := c.GetHeader("x-user-id")
		if uid != "" && !sf.IsAdminManager(uid) {
			examCategorys = services.NewManualService().GetUserCategoryPermissions(uid, examCategorys)
		}
	}

	type tempData struct {
		ID struct {
			ExamCategory      string `bson:"exam_category"`
			ExamChildCategory string `bson:"exam_child_category"`
			QuestionReal      int    `bson:"question_real"`
		} `bson:"_id"`
		Count float64 `bson:"count"`
	}
	var tempResp []tempData
	aggregateF := bson.A{bson.M{"$match": bson.M{"status": 5, "exam_category": bson.M{"$ne": ""}}},
		bson.M{"$group": bson.M{"_id": bson.M{"exam_category": "$exam_category", "exam_child_category": "$exam_child_category", "question_real": "$question_real"}, "count": bson.M{"$sum": 1}}}}
	err = sf.DB().Collection("g_interview_questions").Aggregate(aggregateF, &tempResp)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	countMap := make(map[string]int64)
	for _, i := range tempResp {
		countMap[i.ID.ExamCategory+i.ID.ExamChildCategory+strconv.Itoa(i.ID.QuestionReal)] = int64(i.Count)
	}
	var FinalExamCategorys []appresp.ExamCategoryResp
	data, err := json.Marshal(&examCategorys)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	err = json.Unmarshal(data, &FinalExamCategorys)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	for index, examInfo := range FinalExamCategorys {
		for childIndex, examChildInfo := range examInfo.ChildCategory {
			FinalExamCategorys[index].ChildCategory[childIndex].RealCount = countMap[examInfo.Title+examChildInfo.Title+"1"]
			FinalExamCategorys[index].ChildCategory[childIndex].NotRealCount = countMap[examInfo.Title+examChildInfo.Title+"0"]
		}
		FinalExamCategorys[index].RealCount = countMap[examInfo.Title+"1"]
		FinalExamCategorys[index].NotRealCount = countMap[examInfo.Title+"0"]
	}

	sf.Success(FinalExamCategorys, c)
}

// 题分类
func (sf *Category) QuestionCategory(c *gin.Context) {
	var param struct {
		ExamCategory           string `json:"exam_category" binding:"required" msg:"缺少考试类型"`
		ExamChildCategory      string `json:"exam_child_category"`
		QuestionReal           int8   `json:"question_real"`
		OpenCategoryPermission int    `json:"open_category_permission"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	//这里查询完整题分类 并且将题中过滤出来分类下题数量赋值
	var qc models.QuestionCategory
	filter := bson.M{"exam_category": param.ExamCategory}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	err = sf.DB().Collection("question_category").Where(filter).Take(&qc)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resp := make([]models.QuestionCategoryItem, 0)
	if qc.Categorys != nil {
		resp = qc.Categorys
	}
	realResp := make([]models.QuestionCategoryItem, 0)
	notRealResp := make([]models.QuestionCategoryItem, 0)
	for _, i := range qc.Categorys {
		if i.RealQuestionCount > 0 {
			realResp = append(realResp, i)
		}
		if i.NotRealQuestionCount > 0 {
			notRealResp = append(notRealResp, i)
		}
	}
	if param.QuestionReal == 1 {
		resp = realResp
	}
	if param.QuestionReal == 2 {
		resp = notRealResp
	}
	_, exists := c.Get("APP-SOURCE-TYPE")
	if exists {
		var hasReMenXiTi bool
		for _, item := range resp {
			if item.Title == "热点习题" {
				hasReMenXiTi = true
				break
			}
		}
		if !hasReMenXiTi {
			realResp2 := make([]models.QuestionCategoryItem, 0)
			realResp2 = append(realResp2, models.QuestionCategoryItem{Title: "热点习题", QuestionCount: 10, RealQuestionCount: 10, NotRealQuestionCount: 10, ChildCategory: make([]models.QuestionCategoryItem, 0)})
			// 热门习题
			resp = append(realResp2, resp...)
		}
	}

	// 走权限
	if param.OpenCategoryPermission == 1 {
		uid := c.GetHeader("x-user-id")
		if uid != "" && !sf.IsAdminManager(uid) {
			resp = services.NewManualService().GetUserKeypointsPermissions(uid, param.ExamCategory, param.ExamChildCategory, resp)
		}
	}

	sf.Success(resp, c)
}

// 考试分类
func (sf *Category) CategoryPermissionList(c *gin.Context) {
	var err error
	var examCategorys = make([]models.ExamCategory, 0)
	err = sf.DB().Collection("exam_category").Sort("-sort_number").Find(&examCategorys)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	qclist := make([]models.QuestionCategory, 0)
	qcmap := make(map[string]models.QuestionCategory)
	err = sf.DB().Collection("question_category").Find(&qclist)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	for _, category := range qclist {
		qcmap[category.ExamCategory+"-"+category.ExamChildCategory] = category
	}

	var list = make([]managerresp.CategoryPermissionResponse, 0)
	lastEmptyChild := make([]managerresp.CategoryPermissionResponse, 0)

	for _, category := range examCategorys {
		if len(category.ChildCategory) == 0 {
			category.ChildCategory = make([]models.ExamCategoryItem, 1)
		}

		for _, item := range category.ChildCategory {
			value := category.Title
			if item.Title != "" {
				value += "/" + item.Title
			}
			var listItem = managerresp.CategoryPermissionResponse{
				Value: value,
				Label: value,
				Child: make([]managerresp.CategoryPermissionResponse, 0),
			}

			var qc = qcmap[category.Title+"-"+item.Title]
			for _, questionCategoryItem := range qc.Categorys {
				listItem.Child = append(listItem.Child, managerresp.CategoryPermissionResponse{
					Value: listItem.Value + "-" + questionCategoryItem.Title,
					Label: questionCategoryItem.Title,
					Child: lastEmptyChild,
				})
			}
			list = append(list, listItem)
		}
	}

	sf.Success(list, c)
}
