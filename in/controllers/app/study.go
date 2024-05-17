package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/controllers"
	"interview/services"
	"regexp"
)

type Study struct {
	controllers.Controller
}

func (sf *Study) DataPackList(c *gin.Context) {
	var param struct {
		ExamCategory      string `json:"exam_category"`
		ExamChildCategory string `json:"exam_child_category"`
		JobTag            string `json:"job_tag"`
		Keywords          string `json:"keywords"`
		PageIndex         int64  `json:"page_index"`
		PageSize          int64  `json:"page_size"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.Keywords != "" {
		filter["title"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}
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

	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	paperList, paperCount, err := new(services.Study).DataPackList(filter, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"list": paperList, "count": paperCount}, c)
}

func (sf *Study) CourseList(c *gin.Context) {
	var param struct {
		ExamCategory      string `json:"exam_category"`
		ExamChildCategory string `json:"exam_child_category"`
		JobTag            string `json:"job_tag"`
		Year              string `json:"year"`
		CourseType        int8   `json:"course_type"`
		PageIndex         int64  `json:"page_index"`
		PageSize          int64  `json:"page_size"`
		ProvinceCode      string `json:"province_code"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.Year != "" {
		filter["year"] = param.Year
	}
	if param.CourseType != 0 {
		filter["course_type"] = param.CourseType
	}
	if param.ProvinceCode != "" {
		filter["province_code"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.ProvinceCode)}}
	}

	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	paperList, paperCount, err := new(services.Study).CourseList(filter, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"list": paperList, "count": paperCount}, c)
}
