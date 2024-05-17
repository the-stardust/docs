package manager

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/controllers"
	"interview/router/request"
	"interview/services"
	"regexp"
)

type Study struct {
	controllers.Controller
}

func (sf *Study) DataPackList(c *gin.Context) {
	var param struct {
		Id                string `bson:"id" json:"id"`
		Title             string `bson:"title" json:"title"`
		JobTag            string `json:"job_tag" bson:"job_tag"`                         //省份
		ExamCategory      string `json:"exam_category" bson:"exam_category"`             //考试分类
		ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
		Status            int8   `json:"status"`                                         // 1 正常 -1 禁用
		PageIndex         int64  `json:"page_index"`                                     //
		PageSize          int64  `json:"page_size"`                                      //
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.Id != "" {
		filter["_id"] = sf.ObjectID(param.Id)
	}
	if param.Status != 0 {
		filter["status"] = param.Status
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.JobTag != "" {
		filter["job_tag"] = param.JobTag
	}
	if param.Title != "" {
		filter["title"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Title)}}
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	paperList, paperCount, err := new(services.Study).BackendDataPackList(filter, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"list": paperList, "count": paperCount}, c)
}

func (sf *Study) DataPackEdit(c *gin.Context) {
	var err error
	var param request.RecommendDataPackEditRequest
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}

	_, err = new(services.Study).DataPackEdit(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}

func (sf *Study) CourseList(c *gin.Context) {
	var param struct {
		Id                string `bson:"id" json:"id"`
		Title             string `bson:"title" json:"title"`
		Province          string `json:"province" bson:"province"`                       //省份
		ExamCategory      string `json:"exam_category" bson:"exam_category"`             //考试分类
		ExamChildCategory string `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
		Status            int8   `json:"status"`                                         // 1 正常 -1 禁用
		PageIndex         int64  `json:"page_index"`                                     //
		PageSize          int64  `json:"page_size"`                                      //
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.Id != "" {
		filter["_id"] = sf.ObjectID(param.Id)
	}
	if param.Status != 0 {
		filter["status"] = param.Status
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.Province != "" {
		filter["province"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Province)}}
	}
	if param.Title != "" {
		filter["title"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Title)}}
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	paperList, paperCount, err := new(services.Study).BackendCourseList(filter, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"list": paperList, "count": paperCount}, c)
}

func (sf *Study) CourseEdit(c *gin.Context) {
	var err error
	var param request.RecommendCourseEditRequest
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}

	areaJson := new(services.ServicesBase).AreaJson()
	areaArr := make([]map[string]any, 0)
	json.Unmarshal([]byte(areaJson), &areaArr)
	for _, m := range areaArr {
		province := m["ntitle"].(string)
		provinceCode := m["code"].(string)
		for _, pc := range param.ProvinceCode {
			if pc == provinceCode {
				param.Province = append(param.Province, province)
			}
		}
	}
	if common.InArrCommon("100000", param.ProvinceCode) {
		param.Province = append(param.Province, "全国")
	}
	_, err = new(services.Study).CourseEdit(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success("success", c)
}
