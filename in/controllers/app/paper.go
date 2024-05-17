package app

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"regexp"
)

type Paper struct {
	controllers.Controller
}

func (sf *Paper) PaperList(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		Years             []string `json:"years"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		PaperType         int8     `json:"paper_type"` //1真题卷 2模拟卷 3集合卷
		Keywords          string   `json:"keywords"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"status": 1}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.Years) != 0 {
		filter["year"] = bson.M{"$in": param.Years}
	}
	if param.PaperType != 0 {
		filter["paper_type"] = param.PaperType
	}
	if param.Province != "" {
		filter["area.province"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Province)}}
	}
	if param.City != "" {
		filter["area.city"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.City)}}
	}
	if param.District != "" {
		filter["area.district"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.District)}}
	}
	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"title": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"_id": sf.ObjectID(param.Keywords)},
		}
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	paperList, paperCount, err := new(services.Paper).PaperList(filter, offset, limit)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"paper_list": paperList, "paper_count": paperCount}, c)
}

func (sf *Paper) PaperInfo(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		sf.Error(common.CodeInvalidParam, c, "试卷id不允许为空！")
		return
	}

	filter := bson.M{"status": 1, "_id": sf.ObjectID(id)}
	paper, err := new(services.Paper).PaperInfo(filter)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(map[string]interface{}{"paper_info": paper}, c)
}

func (sf *Paper) QuestionsInPaper(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		sf.Error(common.CodeInvalidParam, c, "试卷id不允许为空！")
		return
	}
	filter := bson.M{"status": 1, "_id": sf.ObjectID(id)}
	paper, err := new(services.Paper).PaperInfo(filter)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	questions := make([]models.GQuestion, 0)
	if len(paper.QuestionIds) > 0 {
		questionFilter := bson.M{"status": 5, "_id": bson.M{"$in": sf.ObjectIDs(paper.QuestionIds)}}
		questions, err = new(services.Question).GetQuestions(questionFilter, paper.QuestionIds)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, err.Error())
			return
		}
	}

	sf.Success(map[string]interface{}{"questions": questions, "question_count": len(questions)}, c)
}
