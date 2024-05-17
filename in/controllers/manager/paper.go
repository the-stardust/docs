package manager

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
		PaperType         int8     `json:"paper_type"` //1真题卷 2模拟卷 3集合卷
		Keywords          string   `json:"keywords"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		Status            int8     `json:"status"` // 1 正常 -1 禁用
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.Status != 0 {
		filter["status"] = param.Status
	}
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

	filter := bson.M{"_id": sf.ObjectID(id)}
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

func (sf *Paper) SavePaper(c *gin.Context) {
	var err error
	var param struct {
		Id                string        `json:"paper_id"` // 试卷id
		Title             string        `json:"title"`
		PaperType         int8          `json:"paper_type"`          //1真题卷 2模拟卷 3集合卷
		ExamCategory      string        `json:"exam_category"`       //考试分类
		ExamChildCategory string        `json:"exam_child_category"` //考试子分类
		Year              string        `json:"year"`
		Area              []models.Area `json:"area"`
		QuestionIds       []string      `json:"question_ids"`
		Status            int8          `json:"status"` // 1 正常 -1 禁用
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	if param.PaperType < 1 || param.PaperType > 3 {
		sf.Error(common.CodeInvalidParam, c, "paper_type的值必须为1~3！")
		return
	}

	// 修改
	if param.Id != "" {
		if param.Status != 1 && param.Status != -1 {
			sf.Error(common.CodeInvalidParam, c, "status必须为1或者-1！")
			return
		}
		updateFilter := bson.M{}
		if param.Title != "" {
			updateFilter["title"] = param.Title
		}
		if param.PaperType != 0 {
			updateFilter["paper_type"] = param.PaperType
		}
		if param.ExamCategory != "" {
			updateFilter["exam_category"] = param.ExamCategory
		}
		if param.ExamChildCategory != "" {
			updateFilter["exam_child_category"] = param.ExamChildCategory
		}
		if param.Year != "" {
			updateFilter["year"] = param.Year
		}
		if param.Area != nil {
			updateFilter["area"] = param.Area
		}
		if param.QuestionIds != nil {
			updateFilter["question_ids"] = param.QuestionIds
			updateFilter["question_count"] = len(param.QuestionIds)

		}
		if param.Status != 0 {
			updateFilter["status"] = param.Status
		}
		_, err = sf.DB().Collection("paper").Where(bson.M{"_id": sf.ObjectID(param.Id)}).Update(updateFilter)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, err.Error())
			return
		}

	} else { // 新增
		var paper models.Paper
		managerId := c.GetHeader("x-user-id")
		paper.Title = param.Title
		paper.PaperType = param.PaperType
		paper.ManagerId = managerId
		paper.ManagerName = new(models.Manager).GetManagerName(managerId)
		paper.ExamCategory = param.ExamCategory
		paper.ExamChildCategory = param.ExamChildCategory
		paper.Year = param.Year
		paper.Area = param.Area
		paper.QuestionIds = param.QuestionIds
		paper.QuestionCount = len(param.QuestionIds)
		paper.Status = 1
		paper.ShortId = common.DecimalConversionBelow32(0, 32)
		if len(param.Area) == 0 {
			paper.Area = []models.Area{}
		}
		if len(param.QuestionIds) == 0 {
			paper.QuestionIds = []string{}
		}
		_, err = sf.DB().Collection("paper").Create(&paper)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, err.Error())
			return
		}
	}
	sf.Success("保存成功", c)

}
