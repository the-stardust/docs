package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/ffmt.v1"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/es"
	"interview/helper"
	"interview/models"
	"interview/router/request"
	"interview/services"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/sashabaranov/go-openai"

	"github.com/olahol/melody"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Question struct {
	controllers.Controller
}

type GPTAnalysis struct {
	Idx     int
	Content string
}

var MelodyClient *melody.Melody

// GetInterviewQuestions 查看所有试题
func (sf *Question) QuestionList(c *gin.Context) {
	var param struct {
		ManagerName            string   `json:"manager_name"`
		Keywords               string   `json:"keywords"`
		ExamCategory           string   `json:"exam_category"`
		ExamChildCategory      string   `json:"exam_child_category"`
		QuestionCategory       []string `json:"question_category"`
		Years                  []int    `json:"years"`
		Province               string   `json:"province"`
		City                   string   `json:"city"`
		District               string   `json:"district"`
		PageIndex              int64    `json:"page_index"`
		PageSize               int64    `json:"page_size"`
		Status                 int32    `json:"status"` // 试题状态
		GPTAnswerStatus        int8     `json:"gpt_answer_status"`
		JobTag                 string   `json:"job_tag"` // 岗位标签，如海关、税务局等
		IsOnlyShowSource       bool     `json:"is_only_show_source"`
		QuestionReal           int8     `json:"question_real"`            // 是否真题，0为查全部,1真题，2模拟题
		OpenCategoryPermission int8     `json:"open_category_permission"` //1启用试题分类限制权限
		LogType                int8     `json:"log_type"`
		QuestionContentType    int8     `json:"question_content_type"` // 试题类别，0普通题，1漫画题, -1所有
		SortType               int8     `json:"sort_type"`             // 试题类别，0按创建时间排序，1同面试AI排序
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("x-user-id")
	filter := bson.M{}

	if param.QuestionContentType != -1 {
		filter["question_content_type"] = param.QuestionContentType
	}
	if param.LogType != 0 {
		filter = bson.M{"log_type": param.LogType}
	}
	if param.QuestionReal != 0 {
		if param.QuestionReal == 1 {
			filter["question_real"] = 1
		} else if param.QuestionReal == 2 {
			filter["question_real"] = 0
		}
	}
	if param.JobTag != "" {
		filter["job_tag"] = param.JobTag
	}
	if param.Status != -1 {
		filter["status"] = param.Status
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.Years) > 0 {
		filter["year"] = bson.M{"$in": param.Years}
	}
	if param.GPTAnswerStatus == 1 {
		filter["gpt_answer.content"] = bson.M{"$ne": ""}
	} else if param.GPTAnswerStatus == 2 {
		filter["gpt_answer.content"] = ""
	}
	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"tags": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}}, bson.M{"_id": sf.ObjectID(param.Keywords)},
			bson.M{"question_category": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"gpt_answer.content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"question_source": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}
	//地区
	if param.Province != "" {
		filter["province"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Province)}}
		if param.Province == "其他" {
			filter["province"] = ""
		}
	}
	if param.City != "" {
		filter["city"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.City)}}
	}
	if param.District != "" {
		filter["district"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.District)}}
	}
	if len(param.QuestionCategory) > 0 {
		if param.OpenCategoryPermission == 1 {
			userKeypoints := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, param.QuestionCategory[0], "", 2, "")
			if _, ok := userKeypoints.(string); ok && (userKeypoints.(string) == "not set" || userKeypoints.(string) == "") {
				for i, v := range param.QuestionCategory {
					filter[fmt.Sprintf("question_category.%d", i)] = v
				}
			} else {
				filter["question_category.0"] = userKeypoints
			}
		} else {
			for i, v := range param.QuestionCategory {
				filter[fmt.Sprintf("question_category.%d", i)] = v
			}
		}
	} else {
		// 试题类型权限判断
		if param.OpenCategoryPermission == 1 {
			userKeypoints := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, "", "", 2, "")
			if _, ok := userKeypoints.(string); !ok || (userKeypoints.(string) != "not set" && userKeypoints.(string) != "") {
				filter["question_category.0"] = userKeypoints
			}
		}
	}
	if param.OpenCategoryPermission == 1 {
		exam_category := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, "", "", 1, param.ExamCategory)
		if _, ok := exam_category.(string); !ok || (exam_category.(string) != "not set" && exam_category.(string) != "") {
			filter["exam_category"] = exam_category
		}
	}
	if param.ManagerName != "" {
		var manager models.Manager
		err = sf.DB().Collection("managers").Where(bson.M{"manager_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.ManagerName)}}}).Take(&manager)
		filter["manager_id"] = manager.ManagerId
		if err != nil {
			filter["manager_id"] = "notfound404"
		}
	}
	if param.IsOnlyShowSource {
		filter["question_source"] = bson.M{"$ne": ""}
	}
	totalCount, _ := sf.DB().Collection("g_interview_questions").Where(filter).Count()
	var questions = []models.GQuestion{}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	resultInfo := make(map[string]interface{})

	if param.IsOnlyShowSource {
		var questionsTempList = []interface{}{}
		var tempList = []bson.M{}
		err = sf.DB().Collection("g_interview_questions").Aggregate(bson.A{
			// bson.M{"$sort": bson.D{
			// 	bson.E{Key: "year", Value: -1},
			// 	bson.E{Key: "month", Value: -1},
			// 	bson.E{Key: "day", Value: -1},
			// 	bson.E{Key: "moment", Value: 1}}},
			bson.M{"$match": filter},
			bson.M{"$group": bson.M{"_id": "$question_source", "year": bson.M{"$last": "$year"}, "month": bson.M{"$last": "$month"}, "day": bson.M{"$last": "$day"}, "moment": bson.M{"$last": "$moment"}, "updated_time": bson.M{"$last": "$updated_time"}, "objectID": bson.M{"$last": "$_id"}}},
			bson.M{"$project": bson.M{"updated_time": bson.M{"$toDate": "$updated_time"}, "objectID": "$objectID", "year": "$year", "month": "$month", "day": "$day", "moment": "$moment"}},
			bson.M{"$sort": bson.D{
				bson.E{Key: "year", Value: -1},
				bson.E{Key: "month", Value: -1},
				bson.E{Key: "day", Value: -1},
				bson.E{Key: "moment", Value: 1},
				bson.E{Key: "objectID", Value: 1}}},
			bson.M{"$skip": offset},
			bson.M{"$limit": limit}}, &tempList)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		for _, l := range tempList {
			if vo, ok := l["_id"].(string); ok {
				tempMap := make(map[string]interface{}, 0)
				tempMap["question_source"] = vo
				tempMap["updated_time"] = l["updated_time"].(primitive.DateTime)
				tempMap["objectID"] = l["objectID"].(primitive.ObjectID)

				var previewing = "0"
				previewingStr, _ := helper.RedisGet(fmt.Sprintf("%s%s", rediskey.GPTQuestionPreviewing, l["objectID"].(primitive.ObjectID).Hex()))
				if previewingStr != "" {
					previewing = previewingStr
				}
				tempMap["gpt_preview_status"] = previewing
				questionsTempList = append(questionsTempList, tempMap)
			}
		}
		resultInfo["list"] = questionsTempList
		totalCount, _ = sf.DB().Collection("g_interview_questions").AggregateCount(bson.A{
			bson.M{"$match": filter},
			bson.M{"$group": bson.M{"_id": "$question_source"}},
		})
		resultInfo["count"] = totalCount

	} else {
		sortBy := []string{"-created_time", "-_id"}
		if param.SortType > 0 {
			switch param.SortType {
			case 1:
				sortBy = []string{"-year", "-month", "-day", "+moment", "+objectID"}
			}
		}
		err = sf.DB().Collection("g_interview_questions").Where(filter).Fields(bson.M{"thinking": 0}).Sort(sortBy...).Skip(offset).Limit(limit).Find(&questions)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		// 答题次数和答题人数
		returnMapChan := make(chan map[string]int64, 1)
		questionIDs := []string{}
		for _, q := range questions {
			questionIDs = append(questionIDs, q.Id.Hex())
		}
		type AnswerCount struct {
			QuestionID  string
			Count       int64
			PeopleCount int64
		}
		tempMapChan := make(chan AnswerCount, len(questionIDs))

		go func() {
			tempMap := make(map[string]int64)
			for v := range tempMapChan {
				tempMap[v.QuestionID+"-count"] = v.Count
				tempMap[v.QuestionID+"-people_count"] = v.PeopleCount
			}
			returnMapChan <- tempMap
		}()

		swg := sizedwaitgroup.New(20)
		for _, q := range questionIDs {
			swg.Add()
			go func(wg *sizedwaitgroup.SizedWaitGroup, qID string, chans chan AnswerCount) {
				defer wg.Done()
				countFilter := bson.M{"question_id": qID}
				answerCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(countFilter).Count()
				_, peopleCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(countFilter).Distinct("user_id")
				chans <- AnswerCount{qID, answerCount, int64(peopleCount)}
			}(&swg, q, tempMapChan)

		}
		swg.Wait()
		close(tempMapChan)
		CountMap := <-returnMapChan
		close(returnMapChan)

		tempQuestions := []models.GQuestion{}
		for _, q := range questions {
			q.AnswerCount = CountMap[q.Id.Hex()+"-count"]
			q.PeopleCount = CountMap[q.Id.Hex()+"-people_count"]

			if q.ManagerID != "" {
				q.ManagerName = new(models.Manager).GetManagerName(q.ManagerID)
			}

			var previewing = "0"
			previewingStr, _ := helper.RedisGet(fmt.Sprintf("%s%s", rediskey.GPTQuestionPreviewing, q.Id.Hex()))
			if previewingStr != "" {
				previewing = previewingStr
			}
			q.GPTPreviewStatus = previewing

			tempQuestions = append(tempQuestions, q)
		}
		resultInfo["list"] = tempQuestions
		resultInfo["count"] = totalCount
	}
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(resultInfo, c)

}

func (sf *Question) QuestionSearchList(c *gin.Context) {
	var param struct {
		ManagerName            string   `json:"manager_name"`
		Keywords               string   `json:"keywords"`
		ExamCategory           string   `json:"exam_category"`
		ExamChildCategory      string   `json:"exam_child_category"`
		QuestionCategory       []string `json:"question_category"`
		Years                  []int    `json:"years"`
		Province               string   `json:"province"`
		City                   string   `json:"city"`
		District               string   `json:"district"`
		PageIndex              int64    `json:"page_index"`
		PageSize               int64    `json:"page_size"`
		Status                 int32    `json:"status"` // 试题状态
		GPTAnswerStatus        int8     `json:"gpt_answer_status"`
		JobTag                 string   `json:"job_tag"`                  // 岗位标签，如海关、税务局等
		QuestionReal           int8     `json:"question_real"`            // 是否真题，0为查全部,1真题，2模拟题
		OpenCategoryPermission int8     `json:"open_category_permission"` //1启用试题分类限制权限
		QuestionContentType    int8     `json:"question_content_type"`    // 试题类别，0普通题，1漫画题, -1所有
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	// 权限控制，获取mongo的查询条件然后组装到es的语句
	filterList := make([]elastic.Query, 0) // es过滤语句
	uid := c.GetHeader("x-user-id")
	mongoFindFilter := bson.M{}
	if len(param.QuestionCategory) > 0 {
		if param.OpenCategoryPermission == 1 {
			userKeypoints := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, param.QuestionCategory[0], "", 2, "")
			if _, ok := userKeypoints.(string); ok && (userKeypoints.(string) == "not set" || userKeypoints.(string) == "") {
				for i, v := range param.QuestionCategory {
					mongoFindFilter[fmt.Sprintf("question_category.%d", i)] = v
				}
			} else {
				mongoFindFilter["question_category.0"] = userKeypoints
			}
		} else {
			for i, v := range param.QuestionCategory {
				mongoFindFilter[fmt.Sprintf("question_category.%d", i)] = v
			}
		}
	} else {
		// 试题类型权限判断
		if param.OpenCategoryPermission == 1 {
			userKeypoints := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, "", "", 2, "")
			if _, ok := userKeypoints.(string); !ok || (userKeypoints.(string) != "not set" && userKeypoints.(string) != "") {
				mongoFindFilter["question_category.0"] = userKeypoints
				// mongoFindFilter["question_category.0"] = userKeypoints
			}
		}
	}
	if param.OpenCategoryPermission == 1 {
		exam_category := sf.UserCategoryPermissionFilter(uid, param.ExamCategory, param.ExamChildCategory, "", "", 1, param.ExamCategory)
		if _, ok := exam_category.(string); !ok || (exam_category.(string) != "not set" && exam_category.(string) != "") {
			if exam_category != "no permission" && exam_category != "" {
				mongoFindFilter["exam_category"] = exam_category
			}
		}
	}
	ffmt.Puts("mongoFindfilter:", mongoFindFilter)
	// 已经获取到mongo的权限控制语句，开始组装
	tempShouldQuery := elastic.NewBoolQuery()
	for key, value := range mongoFindFilter {
		key = strings.Split(key, ".")[0] // 把question_category.0变question_category
		// 此时的value是map[$in:[社会现象 态度观点 社会现象]]这种格式
		if v, ok := value.(primitive.M); ok {
			for _, vv := range v {
				abc := make([]interface{}, 0)
				for _, vvv := range vv.([]string) {
					abc = append(abc, vvv)
				}
				tempShouldQuery.Must(elastic.NewTermsQuery(key, abc...))
			}
		} else {
			tempShouldQuery.Must(elastic.NewTermQuery(key, value))
		}
	}
	filterList = append(filterList, tempShouldQuery)
	ffmt.Puts()
	var kBoost, oBoost float64
	if len(param.Keywords) == 0 {
		sf.Error(common.CodeServerBusy, c, "关键词不允许为空！")
		return
	} else if len(param.Keywords) <= 10 {
		kBoost = 3.0
		oBoost = 1.2
	} else {
		kBoost = 3.5
		oBoost = 1.2
	}
	matchBoost := 2.0
	param.Keywords = strings.ReplaceAll(param.Keywords, "的", "")
	//param.Keywords = strings.ReplaceAll(param.Keywords, "了", "")
	query := elastic.NewBoolQuery()

	query.Should(elastic.NewMatchQuery("name", param.Keywords).Boost(oBoost))
	//query.Should(elastic.NewMatchQuery("desc", param.Keywords).Boost(oBoost))
	query.Should(elastic.NewWildcardQuery("name", param.Keywords+"*").Boost(matchBoost))
	//query.Should(elastic.NewWildcardQuery("desc", param.Keywords+"*").Boost(oBoost))
	query.Should(elastic.NewTermQuery("id", param.Keywords).Boost(kBoost))
	query.Should(elastic.NewTermQuery("name", param.Keywords).Boost(kBoost))
	//query.Should(elastic.NewMatchQuery("name_desc", param.Keywords).Boost(oBoost))
	nestedQuery := elastic.NewNestedQuery("name_struct.content", elastic.NewBoolQuery().Must(elastic.NewMatchQuery("name_struct.content.text", param.Keywords)).Boost(oBoost))
	nestedTermQuery := elastic.NewNestedQuery("name_struct.content", elastic.NewBoolQuery().Must(elastic.NewTermQuery("name_struct.content.text", param.Keywords)).Boost(kBoost))
	query.Should(nestedQuery)
	query.Should(nestedTermQuery)
	// 过滤条件，不参与score计算
	if param.QuestionContentType != -1 {
		filterList = append(filterList, elastic.NewTermQuery("question_content_type", param.QuestionContentType))
	}
	if param.ExamCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_category", param.ExamCategory))
	}
	if param.ExamChildCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_child_category", param.ExamChildCategory))
	}
	if param.ManagerName != "" {
		var manager models.Manager
		err = sf.DB().Collection("managers").Where(bson.M{"manager_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.ManagerName)}}}).Take(&manager)
		if err != nil {
			filterList = append(filterList, elastic.NewTermQuery("creator_user_id", "notfound404"))
		} else {
			filterList = append(filterList, elastic.NewTermQuery("creator_user_id", manager.ManagerId))
		}
	}

	if len(param.QuestionCategory) > 0 {
		if len(param.QuestionCategory) == 1 {
			filterList = append(filterList, elastic.NewTermQuery("question_category", param.QuestionCategory[0]))
		} else {
			cBoolQuery := elastic.NewBoolQuery()
			for _, c := range param.QuestionCategory {
				cBoolQuery.Must(elastic.NewTermQuery("question_category", c))
			}
			filterList = append(filterList, cBoolQuery)
		}
	}

	if len(param.Years) > 0 {
		if len(param.Years) == 1 {
			filterList = append(filterList, elastic.NewTermQuery("year", param.Years[0]))
		} else {
			yearBoolQuery := elastic.NewBoolQuery()
			for _, year := range param.Years {
				yearBoolQuery.Should(elastic.NewTermQuery("year", year))
			}
			filterList = append(filterList, yearBoolQuery)
		}
	}
	if param.Province != "" {
		filterList = append(filterList, elastic.NewTermQuery("province", param.Province))
	}
	if param.City != "" {
		filterList = append(filterList, elastic.NewTermQuery("city", param.City))
	}
	if param.District != "" {
		filterList = append(filterList, elastic.NewTermQuery("district", param.District))
	}
	if param.Status != -1 {
		filterList = append(filterList, elastic.NewTermQuery("status", param.Status))
	}
	if param.GPTAnswerStatus == 1 {
		filterList = append(filterList, elastic.NewBoolQuery().MustNot(elastic.NewTermQuery("gpt_answer.content", "")))
	} else if param.GPTAnswerStatus == 2 {
		filterList = append(filterList, elastic.NewTermQuery("gpt_answer.content", ""))
	}
	if param.JobTag != "" {
		filterList = append(filterList, elastic.NewTermQuery("job_tag", param.JobTag))
	}
	if param.QuestionReal != 0 {
		if param.QuestionReal == 1 {
			filterList = append(filterList, elastic.NewTermQuery("question_real", 1))
		} else if param.QuestionReal == 2 {
			filterList = append(filterList, elastic.NewTermQuery("question_real", 0))
		}
	}
	questions := make([]models.GQuestion, 0)
	ESCfg := global.CONFIG.ES
	searchIndex := es.QuestionIndex
	ElasticClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(ESCfg.ElasticUrl),
		elastic.SetBasicAuth(ESCfg.ElasticName, ESCfg.ElasticPwd),
	)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	query.Filter(filterList...)
	ffmt.Puts(query.Source())

	highLightFilter := elastic.NewHighlight()
	highLightFilter = highLightFilter.Fields(elastic.NewHighlighterField("name"), elastic.NewHighlighterField("id"), elastic.NewHighlighterField("desc"), elastic.NewHighlighterField("name_desc"), elastic.NewHighlighterField("name_struct.content.text"))
	highLightFilter.HighlightFilter(true)
	highLightFilter.RequireFieldMatch(true)
	highLightFilter = highLightFilter.PreTags("<font color='blue'>").PostTags("</font>")

	searchResult, err := ElasticClient.Search().
		Index(searchIndex).
		Query(query).
		SortBy(
			elastic.NewFieldSort("_score").Desc(),
		).
		MinScore(4.00).
		From((int(param.PageIndex) - 1) * int(param.PageSize)).
		Size(int(param.PageSize)).
		Highlight(highLightFilter).
		Do(es.Ctx)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	count := searchResult.TotalHits()
	if count > 0 {
		for _, v := range searchResult.Hits.Hits {
			var gQuestion models.GQuestion
			err = json.Unmarshal(v.Source, &gQuestion)
			if err != nil {
				sf.SLogger().Error(err)
			}
			if gQuestion.ManagerID != "" {
				gQuestion.ManagerName = new(models.Manager).GetManagerName(gQuestion.ManagerID)
			}
			questions = append(questions, gQuestion)
		}
	}

	sf.Success(map[string]interface{}{"list": questions, "count": count}, c)
}

// QuestionInfo 查看试题详情
func (sf *Question) QuestionInfo(c *gin.Context) {
	QuestionID := c.Query("question_id")
	if QuestionID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var question models.GQuestion
	filter := bson.M{"_id": sf.ObjectID(QuestionID)}
	err := sf.DB().Collection("g_interview_questions").Where(filter).Take(&question)
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

	if question.ManagerID != "" {
		question.ManagerName = new(models.Manager).GetManagerName(question.ManagerID)
	}

	// 预生成的答题思路和得分点
	cacheKey := fmt.Sprintf("%s%s", rediskey.GPTQuestionPreview, QuestionID)
	str, err := helper.RedisGet(cacheKey)
	previewItem := make(map[string][]string, 0)
	if str != "" {
		err = json.Unmarshal([]byte(str), &previewItem)
		if err != nil {
			sf.SLogger().Error(err)
		}
	}
	previewIdeas := make([]string, 0)
	previewStandardAnswer := make([]string, 0)
	if _, ok := previewItem["1"]; ok {
		previewIdeas = previewItem["1"]
	}
	if _, ok := previewItem["3"]; ok {
		previewStandardAnswer = previewItem["3"]
	}
	question.PreviewIdeas = previewIdeas
	question.PreviewStandardAnswer = previewStandardAnswer

	sf.Success(question, c)
}

// QuestionSourceInfo 查看属于同一试题来源下的试题列表
func (sf *Question) QuestionSourceInfo(c *gin.Context) {
	var param struct {
		QuestionSource string `json:"question_source"`
		PageIndex      int64  `json:"page_index"`
		PageSize       int64  `json:"page_size"`
		QuestionReal   int8   `json:"question_real"` // 是否真题，0为查全部,1真题，2模拟题
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	if param.QuestionSource == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var questions = []models.GQuestion{}
	filter := bson.M{"question_source": param.QuestionSource, "status": 5}
	if param.QuestionReal != 0 {
		if param.QuestionReal == 1 {
			filter["question_real"] = 1
		} else if param.QuestionReal == 2 {
			filter["question_real"] = 0
		}
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_interview_questions").Sort("-updated_time").Where(filter).Skip(offset).Limit(limit).Find(&questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	// 我作答过的试题次数
	uid := c.GetHeader("X-XTJ-UID")
	answerCountMap := make(map[int]int64)
	if uid != "" {
		// 添加做题次数
		var idSlice []string
		for _, q := range questions {
			idSlice = append(idSlice, q.Id.Hex())
		}
		lock := sync.Mutex{}
		sw := sizedwaitgroup.New(20)
		for i, qId := range idSlice {
			sw.Add()
			go func(questionIndex int, questionId string, swg *sizedwaitgroup.SizedWaitGroup) {
				defer swg.Done()
				tempCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"user_id": uid, "question_id": questionId, "log_type": 1}).Count()
				lock.Lock()
				answerCountMap[questionIndex] = tempCount
				lock.Unlock()
			}(i, qId, &sw)
		}
		sw.Wait()
	}

	tempQuestions := []models.GQuestion{}
	for index, q := range questions {
		year := ""
		month := ""
		day := ""
		if q.Year != 0 {
			year = fmt.Sprintf("%d年", q.Year)
		}
		if q.Month != 0 {
			month = fmt.Sprintf("%d月", q.Month)
		}
		if q.Day != 0 {
			day = fmt.Sprintf("%d日", q.Day)
		}
		q.Date = fmt.Sprintf("%s%s%s%s", year, month, day, q.Moment)
		q.MyAnswerCount = answerCountMap[index]
		tempQuestions = append(tempQuestions, q)
	}

	result := make(map[string]interface{}, 0)
	result["count"] = len(tempQuestions)
	result["list"] = tempQuestions
	sf.Success(result, c)
}

// SaveInterviewQuestion 保存试题
func (sf *Question) SaveQuestion(c *gin.Context) {
	var param struct {
		QuestionId          string               `json:"question_id"` // 试题ID
		Tags                []string             `json:"tags"`
		Name                string               `json:"name"` // 试题名称
		NameStruct          models.CommonContent `json:"name_struct" bson:"name_struct"`
		NameDesc            string               `json:"name_desc" bson:"name_desc"`
		Desc                string               `json:"desc"`
		Answer              string               `json:"answer"`
		Status              int32                `json:"status"`              // 试题状态
		ExamCategory        string               `json:"exam_category"`       //考试分类
		ExamChildCategory   string               `json:"exam_child_category"` //考试子分类
		QuestionCategory    []string             `json:"question_category"`   //题分类
		GPTAnswer           []models.GPTAnswer   `json:"gpt_answer"`
		Year                int                  `json:"year"`  // 年份
		Month               int                  `json:"month"` // 月份
		Day                 int                  `json:"day"`   // 日
		Province            string               `json:"province"`
		City                string               `json:"city"`
		District            string               `json:"district"`
		JobTag              string               `json:"job_tag"`                                            // 岗位标签，如海关、税务局等
		QuestionSource      string               `json:"question_source"`                                    // 试题来源
		AreaCodes           []string             `json:"area_codes"`                                         // 地区代码
		Moment              string               `json:"moment"`                                             // 上午或者下午
		AnswerTime          int64                `json:"answer_time"`                                        // 答题时间，单位为秒
		QuestionReal        int8                 `json:"question_real"`                                      // 是否为真题，1是，0不是
		QuestionContentType int8                 `json:"question_content_type" bson:"question_content_type"` // 试题类别，0普通题(纯文字），1漫画题（带文字）
		ScorePoint          string               `json:"score_point"`                                        // 评分要点
		ExplainUrl          string               `json:"explain_url"`                                        // 试题讲解
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	uid := c.GetHeader("x-user-id")
	var question models.GQuestion
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	// 去重
	//existId := new(services.Question).GetSimilarStrQuestion("", param.Name) =============
	//returnMsg := ""
	//if len(existId) > 0 {
	//	if param.QuestionId == "" {
	//		returnMsg = fmt.Sprintf("已有相同试题，请勿重复上传！\n重复试题ID：%s", strings.Join(existId, ","))
	//		sf.Error(common.DuplicateQuestion, c, returnMsg)
	//		return
	//	} else {
	//		if !common.InArr(param.QuestionId, existId) {
	//			returnMsg = fmt.Sprintf("已有相同试题，请勿重复保存！\n重复试题ID：%s", strings.Join(existId, ","))
	//			sf.Error(common.DuplicateQuestion, c, returnMsg)
	//			return
	//		}
	//	}
	//}
	//oldHash := ""
	//newHash := ""
	//insertedId := "" ====================
	// 如果存在试题ID，代表是修改操作
	if param.QuestionId != "" {
		questionFilter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
		err = sf.DB().Collection("g_interview_questions").Where(questionFilter).Take(&question)
		if err != nil {
			if sf.MongoNoResult(err) {
				sf.Error(common.CodeServerBusy, c, "修改失败,试题不存在!")
				return
			} else {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
		}
		if param.Status == 5 && param.Name != question.Name {
			question.TTSUrl = models.TTSUrl{}
		}
		// 如果修改了题目内容，处理去重
		//oldHash = services.StrSimHash(question.Name)   ======
		//newHash = services.StrSimHash(param.Name)  ========
		question.Moment = param.Moment
		question.Name = param.Name
		question.Answer = param.Answer
		question.Status = param.Status
		question.Desc = param.Desc
		question.Tags = param.Tags
		question.GPTAnswer = param.GPTAnswer
		question.ExamCategory = param.ExamCategory
		question.ExamChildCategory = param.ExamChildCategory
		question.QuestionCategory = param.QuestionCategory
		question.Province = param.Province
		question.City = param.City
		question.District = param.District
		question.Year = param.Year
		question.Month = param.Month
		question.Day = param.Day
		question.JobTag = param.JobTag
		question.QuestionSource = param.QuestionSource
		question.AreaCodes = param.AreaCodes
		question.AnswerTime = param.AnswerTime
		question.QuestionReal = param.QuestionReal
		question.QuestionContentType = param.QuestionContentType
		question.NameStruct = param.NameStruct
		question.NameDesc = param.NameDesc
		question.ScorePoint = param.ScorePoint
		question.ExplainUrl = param.ExplainUrl
		// question.CategoryId = param.CategoryId
		question.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
		err = sf.DB().Collection("g_interview_questions").Save(&question)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		// 从试卷中删除下架的试题
		if param.Status == 0 || param.Status == 9 {
			var bulkPullSlice []mongo.WriteModel
			var bulkSubSlice []mongo.WriteModel
			var papers []models.Paper
			err = sf.DB().Collection("paper").Where(bson.M{"question_ids": bson.M{"$in": []string{param.QuestionId}}}).Find(&papers)
			if err != nil {
				sf.SLogger().Error(err)
			}
			for _, paper := range papers {
				bulkPullSlice = append(bulkPullSlice, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": paper.Id}).SetUpdate(bson.M{"$pull": bson.M{"question_ids": param.QuestionId}}))
				bulkSubSlice = append(bulkSubSlice, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": paper.Id}).SetUpdate(bson.M{"$set": bson.M{"question_count": len(paper.QuestionIds) - 1}}))
			}
			if len(bulkPullSlice) > 0 {
				_, err = sf.DB().Collection("paper").BulkWrite(bulkPullSlice)
				if err != nil {
					sf.SLogger().Error(err)
				}
			}

			if len(bulkSubSlice) > 0 {
				_, err = sf.DB().Collection("paper").BulkWrite(bulkSubSlice)
				if err != nil {
					sf.SLogger().Error(err)
				}
			}
		}
	} else {
		// 新增试题
		// newHash = services.StrSimHash(param.Name)  ==========
		question.Moment = param.Moment
		question.Name = param.Name
		question.Answer = param.Answer
		question.Status = param.Status
		question.Desc = param.Desc
		question.Tags = param.Tags
		question.CreatorUserId = uid
		question.GPTAnswer = param.GPTAnswer
		question.ExamCategory = param.ExamCategory
		question.ExamChildCategory = param.ExamChildCategory
		question.QuestionCategory = param.QuestionCategory
		question.Province = param.Province
		question.City = param.City
		question.District = param.District
		question.Year = param.Year
		question.Month = param.Month
		question.Day = param.Day
		question.JobTag = param.JobTag
		question.QuestionSource = param.QuestionSource
		question.AreaCodes = param.AreaCodes
		question.ManagerID = uid
		question.AnswerTime = param.AnswerTime
		question.QuestionReal = param.QuestionReal
		question.QuestionContentType = param.QuestionContentType
		question.NameStruct = param.NameStruct
		question.NameDesc = param.NameDesc
		question.ScorePoint = param.ScorePoint
		question.ExplainUrl = param.ExplainUrl
		_, err = sf.DB().Collection("g_interview_questions").Create(&question)
		// createResp, err := sf.DB().Collection("g_interview_questions").Create(&question) ========
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		// insertedId = createResp.InsertedID.(primitive.ObjectID).Hex() ======
	}
	//go func() {
	//	defer func() {
	//		if recoverError := recover(); recoverError != nil {
	//			sf.SLogger().Error("build home keypoint panic:", recoverError)
	//		}
	//	}()
	//	services.InitInterviewQuestionService().BuildKeypoint()
	//}()
	// es中保存
	tempEsClient, err := es.CreateEsClient()
	if err == nil {
		err = es.QuestionsAddToEs(tempEsClient, context.Background(), []models.GQuestion{question})
		if err != nil {
			sf.SLogger().Error(err)
		}
	} else {
		sf.SLogger().Error(err)
	}

	rdb.Do("HSET", rediskey.InterviewGPTQuestionId2Name, question.Id.Hex(), question.GetWantedQuestionContent())

	// 删除掉hash值对应存储的id
	//if oldHash != "" { ==============
	//	qh := new(models.QuestionHash)
	//	err = sf.DB().Collection("question_hash").Where(bson.M{"hash": oldHash}).Take(qh)
	//	if err != nil {
	//		sf.SLogger().Error(err)
	//	} else {
	//		newIds := make([]string, 0)
	//		for _, i := range qh.QuestionIds {
	//			if i != param.QuestionId {
	//				newIds = append(newIds, i)
	//			}
	//		}
	//		qh.QuestionIds = newIds
	//		err = sf.DB().Collection("question_hash").Where(bson.M{"hash": oldHash}).Save(qh)
	//		if err != nil {
	//			sf.SLogger().Error(err)
	//		}
	//	}
	//}
	//// 添加新hash信息
	//if newHash != "" {
	//	if param.QuestionId != "" {
	//		insertedId = param.QuestionId
	//	}
	//	qh := new(models.QuestionHash)
	//	err = sf.DB().Collection("question_hash").Where(bson.M{"hash": newHash}).Take(qh)
	//	if err == nil {
	//		qh.QuestionIds = append(qh.QuestionIds, insertedId)
	//		qh.QuestionIds = common.RemoveDuplicateElement(qh.QuestionIds)
	//		qh.QuestionIdsTotal = len(qh.QuestionIds)
	//		_, err = sf.DB().Collection("question_hash").Where(bson.M{"hash": newHash}).Update(qh)
	//		if err != nil {
	//			sf.SLogger().Error(err)
	//		}
	//	} else {
	//		if sf.MongoNoResult(err) {
	//			qh.Hash = newHash
	//			qh.QuestionIds = make([]string, 0)
	//			qh.QuestionIds = append(qh.QuestionIds, insertedId)
	//			qh.QuestionIdsTotal = len(qh.QuestionIds)
	//			_, err = sf.DB().Collection("question_hash").Where(bson.M{"hash": newHash}).Create(qh)
	//			if err != nil {
	//				sf.SLogger().Error(err)
	//			}
	//		} else {
	//			sf.SLogger().Error(err)
	//		}
	//	}
	//} ===============

	sf.Success(nil, c)
}
func (sf *Question) MakeGPTAnswer(c *gin.Context) {
	var err error
	var param struct {
		QuestionName   string  `json:"question_name" binding:"required" msg:"invalid question_name"` // 试题内容
		SystemContent  string  `json:"system_content"`                                               //系统信息
		Prompt         string  `json:"prompt"`                                                       //提问内容
		Temperature    float32 `json:"temperature"`                                                  //温度
		TopP           float32 `json:"top_p"`
		MakeCount      int     `json:"make_count"`      //生成数量
		QuestionAnswer string  `json:"question_answer"` // 学员回答内容
		IsComment      bool    `json:"is_comment"`      // 是否为点评
		IsUse4         bool    `json:"is_use_4"`        // 是否使用GPT4版本
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	if param.MakeCount == 0 {
		param.MakeCount = 1
	}
	if param.MakeCount > 10 {
		param.MakeCount = 10
	}
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	systemContent := param.SystemContent
	if systemContent == "" {
		systemContent = "你是一名负责公务员考试中面试环节的老师."
	}

	prompt := ""
	// 带占位符的prompt
	if strings.Contains(param.Prompt, "%s") || strings.Contains(param.Prompt, "%+v") {
		if param.Prompt != "" && param.IsComment == true {
			prompt = fmt.Sprintf(param.Prompt, param.QuestionName, param.QuestionAnswer)
		} else if param.Prompt != "" {
			prompt = fmt.Sprintf(param.Prompt, param.QuestionName)
		} else {
			prompt = fmt.Sprintf(`  面试题目会放在【】里。
		对于【%+v】，请提供一些提示，帮助学生组织好他们的回答。 `, param.QuestionName)
		}
	} else { // 不带占位符的prompt
		if param.Prompt != "" && param.IsComment {
			prompt = param.Prompt + "\n【题目】：" + param.QuestionName + "\n【学生答案】：" + param.QuestionAnswer
		} else if param.Prompt != "" {
			prompt = param.Prompt + "\n【题目】：" + param.QuestionName
		} else {
			prompt = "面试题目会放在【题目】里。对于【题目】，请提供一些提示，帮助学生组织好他们的回答。" + "\n【题目】：" + param.QuestionName
		}
	}
	chatCompletionMessage := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemContent,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}
	var temperature float32 = 0
	if param.Temperature != 0 {
		temperature = param.Temperature
	}
	var topP float32 = 0.95
	if param.Temperature != 0 {
		topP = param.TopP
	}
	// 同时生成三条

	analysisCh := make(chan GPTAnalysis, param.MakeCount)
	respCh := make(chan []GPTAnalysis, param.MakeCount)
	go func() {
		var analysisT = []GPTAnalysis{}
		for analysis := range analysisCh {
			analysisT = append(analysisT, analysis)
		}
		respCh <- analysisT
	}()
	var sw = sizedwaitgroup.New(10)
	for i := 0; i < param.MakeCount; i++ {
		sw.Add()
		go func(idx int, temperature float32, topP float32, SystemContent string, Prompt string, ch chan GPTAnalysis, swg *sizedwaitgroup.SizedWaitGroup) {
			defer swg.Done()
			resp, err := new(services.GPT).MakeAnswer1(temperature, topP, chatCompletionMessage, 1, param.IsUse4)
			if err != nil {
				sf.SLogger().Error(err, param.IsUse4)
				ch <- GPTAnalysis{Idx: idx, Content: "出错了，暂时无法生成"}
			} else {
				ch <- GPTAnalysis{Idx: idx, Content: resp[0]}
			}
		}(i, temperature, topP, param.SystemContent, param.Prompt, analysisCh, &sw)
	}
	sw.Wait()
	close(analysisCh)
	GPTAnalysisArr := <-respCh
	close(respCh)
	sort.Slice(GPTAnalysisArr, func(i, j int) bool {
		return GPTAnalysisArr[i].Idx < GPTAnalysisArr[j].Idx
	})
	GPTAnalysis := []string{}
	for _, v := range GPTAnalysisArr {
		GPTAnalysis = append(GPTAnalysis, v.Content)
	}
	sf.Success(GPTAnalysis, c)

	//answers, err := new(services.GPT).MakeAnswer1(temperature, topP, chatCompletionMessage, 1)
	//if err == nil {
	//	if len(answers) > 0 {
	//		sf.Success(map[string]interface{}{"content": answers[0]}, c)
	//	} else {
	//		sf.SLogger().Errorf("gpt结果异常:%+v", answers)
	//		sf.Error(common.CodeServerBusy, c)
	//	}
	//
	//} else {
	//	sf.SLogger().Error(err)
	//	sf.Error(common.CodeServerBusy, c)
	//}

}

func (sf *Question) AnswerList(c *gin.Context) {
	var err error
	var param struct {
		QuestionID        string   `json:"question_id"`
		Keywords          string   `json:"keywords"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		PracticeType      int8     `json:"practice_type"` // 练习种类，0是全部、11是看题-普通模式，12是看题-对镜模式，13是看题-考官模式 21是听题-普通模式，22是听题-对镜模式，23是听题-考官模式
		StartTime         string   `json:"start_time"`
		EndTime           string   `json:"end_time"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		LogType           int8     `json:"log_type"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	// param.StartTime = sf.PerfectTimeFormat(param.StartTime, 1)
	// param.EndTime = sf.PerfectTimeFormat(param.EndTime, 2)
	filter := bson.M{}
	if param.LogType != 0 {
		filter["log_type"] = param.LogType
	}
	if param.Province != "" {
		filter["province"] = param.Province
	}
	if param.City != "" {
		filter["city"] = param.City
	}
	if param.District != "" {
		filter["district"] = param.District
	}
	if param.StartTime != "" && param.EndTime != "" {
		filter["created_time"] = bson.M{"$gte": param.StartTime, "$lte": param.EndTime}
	}
	if param.PracticeType != 0 {
		filter["practice_type"] = param.PracticeType
	}
	if param.QuestionID != "" {
		filter["question_id"] = param.QuestionID
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.QuestionCategory) != 0 {
		for i, q := range param.QuestionCategory {
			filter[fmt.Sprintf("question_category.%d", i)] = q
		}
	}

	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"user_id": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"_id": param.Keywords},
			bson.M{"question_id": param.Keywords},
			bson.M{"answer.0.voice_text": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"gpt_comment.content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"question_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}
	var answerLogs []models.GAnswerLog
	var totalCount int64
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Skip(offset).Limit(limit).Sort("-updated_time").Find(&answerLogs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务器繁忙，请稍后再试")
		return
	}
	totalCount, err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "服务器繁忙，请稍后再试")
		return
	}
	//}

	// 增加头像和昵称
	list := []models.GAnswerLog{}
	ids := []string{}
	for _, i := range answerLogs {
		ids = append(ids, i.UserId)
	}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo(ids, appCode, 1)
	for _, i := range answerLogs {
		i.UserName = userMap[i.UserId].Nickname
		i.Avatar = userMap[i.UserId].Avatar
		if i.QuestionCategory == nil {
			i.QuestionCategory = []string{}
		}
		list = append(list, i)
	}

	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

func (sf *Question) AnswerListWithES(c *gin.Context) {
	var err error
	var param struct {
		QuestionID        string   `json:"question_id"`
		Keywords          string   `json:"keywords"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		PracticeType      int8     `json:"practice_type"` // 练习种类，0是全部、11是看题-普通模式，12是看题-对镜模式，13是看题-考官模式 21是听题-普通模式，22是听题-对镜模式，23是听题-考官模式
		StartTime         string   `json:"start_time"`
		EndTime           string   `json:"end_time"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		LogType           int8     `json:"log_type"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	var kBoost, oBoost float64
	if len(param.Keywords) == 0 {
		sf.Error(common.CodeServerBusy, c, "关键词不允许为空！")
		return
	} else if len(param.Keywords) <= 10 {
		kBoost = 3.0
		oBoost = 1.2
	} else {
		kBoost = 3.5
		oBoost = 1.2
	}
	param.Keywords = strings.ReplaceAll(param.Keywords, "的", "")
	param.Keywords = strings.ReplaceAll(param.Keywords, "了", "")
	query := elastic.NewBoolQuery()
	filterList := make([]elastic.Query, 0)

	if param.LogType != 0 {
		filterList = append(filterList, elastic.NewTermQuery("log_type", param.LogType))
	}
	if param.Province != "" {
		filterList = append(filterList, elastic.NewTermQuery("province", param.Province))
	}
	if param.City != "" {
		filterList = append(filterList, elastic.NewTermQuery("city", param.City))
	}
	if param.District != "" {
		filterList = append(filterList, elastic.NewTermQuery("district", param.District))
	}
	if param.StartTime != "" && param.EndTime != "" {
		filterList = append(filterList, elastic.NewRangeQuery("created_time").Gte(param.StartTime).Lte(param.EndTime).Boost(kBoost))
	}
	if param.PracticeType != 0 {
		filterList = append(filterList, elastic.NewTermQuery("practice_type", param.PracticeType))
	}
	if param.QuestionID != "" {
		filterList = append(filterList, elastic.NewTermQuery("question_id", param.QuestionID))
	}
	if param.ExamCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_category", param.ExamCategory))
	}
	if param.ExamChildCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_child_category", param.ExamChildCategory))
	}
	if len(param.QuestionCategory) != 0 {
		for _, q := range param.QuestionCategory {
			filterList = append(filterList, elastic.NewTermQuery("question_category", q))
		}
	}

	query.Should(elastic.NewMatchQuery("question_name", param.Keywords).Boost(kBoost))
	//query.Should(elastic.NewMatchQuery("answer.0.voice_text", param.Keywords).Boost(kBoost))
	//query.Should(elastic.NewMatchQuery("gpt_comment.content", param.Keywords).Boost(kBoost))
	query.Should(elastic.NewTermQuery("id", param.Keywords).Boost(oBoost))
	query.Should(elastic.NewTermQuery("user_id", param.Keywords).Boost(kBoost))
	query.Should(elastic.NewTermQuery("question_id", param.Keywords).Boost(oBoost))
	nestedQuery := elastic.NewNestedQuery("name_struct.content", elastic.NewBoolQuery().Must(elastic.NewMatchQuery("name_struct.content.text", param.Keywords)).Boost(kBoost))
	query.Should(nestedQuery)

	ESCfg := global.CONFIG.ES
	searchIndex := es.AnswerLogIndex
	ElasticClient, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(ESCfg.ElasticUrl),
		elastic.SetBasicAuth(ESCfg.ElasticName, ESCfg.ElasticPwd),
	)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	query.Filter(filterList...)
	ffmt.Puts(query.Source())
	searchResult, err := ElasticClient.Search().
		Index(searchIndex).
		Query(query).
		SortBy(
			elastic.NewFieldSort("_score").Desc(),
			elastic.NewFieldSort("created_time").Desc(),
			//elastic.NewFieldSort("id").Desc(),
		).
		MinScore(4.00).
		From(int((param.PageIndex - 1) * param.PageSize)).
		Size(int(param.PageSize)).
		Do(es.Ctx)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	totalCount := searchResult.TotalHits()
	returnList := make([]models.GAnswerLog, 0)
	if totalCount > 0 {
		for _, v := range searchResult.Hits.Hits {
			var gAnswerLog models.GAnswerLog
			err = json.Unmarshal(v.Source, &gAnswerLog)
			if err != nil {
				sf.SLogger().Error(err)
			}
			returnList = append(returnList, gAnswerLog)
		}
	}

	// 增加头像和昵称
	list := []models.GAnswerLog{}
	ids := []string{}
	for _, i := range returnList {
		ids = append(ids, i.UserId)
	}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo(ids, appCode, 1)
	for _, i := range returnList {
		i.UserName = userMap[i.UserId].Nickname
		i.Avatar = userMap[i.UserId].Avatar
		if i.QuestionCategory == nil {
			i.QuestionCategory = []string{}
		}
		list = append(list, i)
	}

	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

// CustomQuestionList 查看用户自己提问的所有试题
func (sf *Question) CustomQuestionList(c *gin.Context) {
	var param struct {
		UserID            string `json:"user_id"`
		Keywords          string `json:"keywords"`
		PageIndex         int64  `json:"page_index"`
		PageSize          int64  `json:"page_size"`
		AnswerType        int8   `json:"answer_type"`
		ExamCategory      string `json:"exam_category"`
		ExamChildCategory string `json:"exam_child_category"` //考试子分类
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{}
	if param.AnswerType != 0 {
		filter["answer_type"] = param.AnswerType
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.UserID != "" {
		filter["user_id"] = param.UserID
	}
	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"_id": param.Keywords},
			bson.M{"gpt_answer.content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
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
	// 增加头像和昵称
	list := []models.GCustomQuestion{}
	ids := []string{}
	for _, i := range questions {
		ids = append(ids, i.UserId)
	}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo(ids, appCode, 1)
	for _, i := range questions {
		i.UserName = userMap[i.UserId].Nickname
		i.UserAvatar = userMap[i.UserId].Avatar
		if i.QuestionCategory == nil {
			i.QuestionCategory = []string{}
		}
		list = append(list, i)
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = list
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

func (sf *Question) PromptFromCategory(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   // 试题分类
		AnswerType        int8     `json:"answer_type"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	if len(param.QuestionCategory) < 1 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, "question_category不能为空")
		return
	}

	filter := bson.M{"answer_type": param.AnswerType}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}

	categoryList := [][]string{}
	categoryCount := len(param.QuestionCategory)
	for index, _ := range param.QuestionCategory {
		endIndex := categoryCount - index
		if endIndex == 0 {
			break
		}
		categoryList = append(categoryList, param.QuestionCategory[0:endIndex])
	}
	var t models.CategoryGPTPrompt
	for _, category := range categoryList {
		filter["question_category"] = category
		err = sf.DB().Collection("category_gpt_prompt").Where(filter).Take(&t)
		if err == nil {
			break
		}
	}
	sf.Success(map[string]interface{}{"system_content": t.SystemContent, "prompt": t.Prompt, "temperature": t.Temperature, "top_p": t.TopP}, c)
}

func (sf *Question) PromptFromCategoryTemp(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   // 试题分类

	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	if len(param.QuestionCategory) < 1 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, "question_category不能为空")
		return
	}

	filter := bson.M{"answer_type": 3}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	orFilter := bson.A{}
	for _, q := range param.QuestionCategory {
		orFilter = append(orFilter, bson.M{"question_category": q})
	}
	filter["$or"] = orFilter
	var t models.CategoryGPTPrompt
	err = sf.DB().Collection("category_gpt_prompt").Where(filter).Take(&t)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Success(map[string]interface{}{"system_content": "", "prompt": "", "temperature": 0.0, "top_p": 0.0}, c)
			return
		}
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success(map[string]interface{}{"system_content": t.SystemContent, "prompt": t.Prompt, "temperature": t.Temperature, "top_p": t.TopP}, c)

}

// SavePrompt 根据考试科目和试题分类设置GPT提示信息
func (sf *Question) SavePrompt(c *gin.Context) {
	var param struct {
		AnswerType        int8     `json:"answer_type"`
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   // 试题分类
		SystemContent     string   `json:"system_content"`      //系统信息
		Prompt            string   `json:"prompt"`              //提问内容
		Temperature       float64  `json:"temperature"`
		TopP              float64  `json:"top_p"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	if len(param.QuestionCategory) < 1 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, "question_category不能为空")
		return
	}
	if param.SystemContent == "" || param.Prompt == "" {
		sf.Error(common.CodeInvalidParam, c, "系统信息或提问内容不能为空")
		return
	}

	filter := bson.M{"answer_type": param.AnswerType, "question_category": param.QuestionCategory, "exam_category": param.ExamCategory, "exam_child_category": param.ExamChildCategory}
	var t models.CategoryGPTPrompt
	err = sf.DB().Collection("category_gpt_prompt").Where(filter).Take(&t)
	if err != nil {
		if sf.MongoNoResult(err) {
			// 新增
			t.ExamCategory = param.ExamCategory
			t.ExamChildCategory = param.ExamChildCategory
			t.SystemContent = param.SystemContent
			t.Prompt = param.Prompt
			t.Temperature = param.Temperature
			t.TopP = param.TopP
			t.AnswerType = param.AnswerType
			t.QuestionCategory = param.QuestionCategory
			_, err = sf.DB().Collection("category_gpt_prompt").Create(&t)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
				return
			}
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	} else {
		// 保存
		t.ExamCategory = param.ExamCategory
		t.ExamChildCategory = param.ExamChildCategory
		t.SystemContent = param.SystemContent
		t.Prompt = param.Prompt
		t.Temperature = param.Temperature
		t.TopP = param.TopP
		t.AnswerType = param.AnswerType
		t.QuestionCategory = param.QuestionCategory
		err = sf.DB().Collection("category_gpt_prompt").Save(&t)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "服务繁忙，请稍后再试")
			return
		}
	}

	sf.Success(nil, c)

}

func (sf *Question) QuestionAnswerLog(c *gin.Context) {
	var param struct {
		QuestionID string `json:"question_id"`
		PageIndex  int64  `json:"page_index"`
		PageSize   int64  `json:"page_size"`
		LogType    int8   `json:"log_type"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	filter := bson.M{"question_id": param.QuestionID}
	if param.LogType != 0 {
		filter["log_type"] = param.LogType
	}
	totalCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(filter).Count()
	var answerLogs = []models.GAnswerLog{}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Sort("-created_time").Skip(offset).Limit(limit).Find(&answerLogs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	tempLogs := []string{}
	for _, a := range answerLogs {
		if len(a.Answer) > 0 {
			answer := a.Answer[0].VoiceText
			tempLogs = append(tempLogs, answer)
		} else {
			tempLogs = append(tempLogs, "此记录没有有效回答")
		}
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = tempLogs
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

type commonMessage struct {
	Temperature   float32           `json:"temperature"`
	TopP          float32           `json:"top_p"`
	SystemContent string            `json:"system_content"`
	Prompt        string            `json:"prompt"`
	Extra         map[string]string `json:"extra"`    // to_topic需要用到的一些额外信息
	ToTopic       string            `json:"to_topic"` // 生成好的内容推送到的队列
}

// 预生成
func (sf *Question) GPTPreview(c *gin.Context) {
	var err error
	var param = new(request.GPTPreviewRequest)
	err = c.ShouldBindBodyWith(param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	questionIds := sf.ObjectIDs(param.Ids)
	questions := make([]models.GQuestion, 0)
	err = sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": bson.M{"$in": questionIds}}).Find(&questions)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	go func() {
		var defaultPrompt = map[int8]map[string]string{
			1: { // 答题思路
				"times":  "3",
				"system": "你是一名负责公务员考试中面试环节的老师。",
				"prompt": "面试题目会放在【】里。对于【%+v】，请提供一些提示，帮助学生组织好他们的回答。",
			},
			3: { // 标准回答
				"times":  "3",
				"system": "你是一名参加公务员面试的考生。",
				"prompt": "面试题目会放在【】里。对于【%+v】，请思考后以第一人称的视角提供标准答案，以便学生在答题时可以参考你的回答。",
			},
		}
		var msgs = make([]commonMessage, 0)
		for _, question := range questions {
			helper.RedisSet(fmt.Sprintf("%s%s", rediskey.GPTQuestionPreviewing, question.Id.Hex()), "1", 6*86400)
			for answerType, dp := range defaultPrompt {
				prompt := fmt.Sprintf(dp["prompt"], question.GetWantedQuestionContent())
				// 生成的次数
				makeNum, _ := strconv.Atoi(dp["times"])
				for i := 0; i < makeNum; i++ {
					msgs = append(msgs, commonMessage{
						SystemContent: dp["system"],
						Prompt:        prompt,
						Temperature:   0,
						TopP:          0.95,
						Extra: map[string]string{
							"paper_id":     "",
							"question_id":  question.Id.Hex(),
							"preview_type": strconv.Itoa(int(answerType)),
							"from":         "interview",
						},
						ToTopic: global.CONFIG.Kafka.GPTQuestionPreviewTopic,
					})
				}
			}
		}

		producer := helper.KafkaNewWriter()
		defer producer.Close()
		for _, msg := range msgs {
			btData, err := json.Marshal(msg)
			if err != nil {
				sf.SLogger().Error("json.marshal fail, err:", err)
				continue
			}

			err = helper.KafkaSendMassage(producer, global.CONFIG.Kafka.QuestionGPTCommonTopic, btData)
			if err != nil {
				sf.SLogger().Error("KafkaSendMassage fail, err:", err)
			}
		}
	}()

	sf.Success("success", c)
}

func (sf *Question) QuestionLogAreas(c *gin.Context) {
	var err error
	var param struct {
		ExamCategory      string `json:"exam_category"`
		ExamChildCategory string `json:"exam_child_category"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	// 先从redis查询地区数据
	isNeedQueryFromMongo := false
	categoryName := param.ExamCategory + param.ExamChildCategory
	var areas []models.TreeNode
	var locations []models.Location
	resp, err := helper.RedisHGet(string(rediskey.QuestionAreas), categoryName)
	if err != nil || resp == "" {
		sf.SLogger().Error(err)
		isNeedQueryFromMongo = true
	}
	err = json.Unmarshal([]byte(resp), &areas)
	if err != nil {
		sf.SLogger().Error(err)
		isNeedQueryFromMongo = true
	}

	if isNeedQueryFromMongo {
		// 筛选出省市区字段存在且不为空的
		filter := bson.M{"$and": bson.A{
			bson.M{"province": bson.M{"$exists": true, "$ne": ""}},
		}}
		if param.ExamCategory != "" {
			filter["exam_category"] = param.ExamCategory
		}
		if param.ExamChildCategory != "" {
			filter["exam_child_category"] = param.ExamChildCategory
		}
		filter["log_type"] = 1
		aggregateFilter := bson.A{
			bson.M{"$match": filter},
			// 按省市区分组
			bson.M{"$group": bson.M{"_id": bson.M{
				"province": "$province",
				"city":     "$city",
				"district": "$district",
			}}},
			bson.M{
				"$project": bson.M{
					"_id":      0,
					"province": "$_id.province",
					"city":     "$_id.city",
					"district": "$_id.district",
				}},
		}
		err = sf.DB().Collection("g_interview_answer_logs").Aggregate(aggregateFilter, &locations)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}

		root := new(models.TreeNode).BuildTree(locations)
		areas = root.Child

		areaData, err := json.Marshal(areas)
		if err != nil {
			sf.SLogger().Error(err)
		} else {
			err = helper.RedisHSet(string(rediskey.QuestionAreas), categoryName, areaData)
			if err != nil {
				sf.SLogger().Error(err)
			}
			err = helper.RedisEXPIRE(string(rediskey.QuestionAreas), 60*30)
			if err != nil {
				sf.SLogger().Error(err)
			}
		}
	}

	sf.Success(areas, c)
}

//func (sf *Question) Qc(c *gin.Context) {
//	err := sf.Do(sf.ObjectID("64a28200ab858935259ce28f"))
//	sf.Success(err, c)
//}
//
//func (sf *Question) Do(objectID interface{}) error {
//	sf.SLogger().Info("启动各分类试题计数任务")
//	var err error
//	var qc = models.QuestionCategory{}
//	err = sf.DB().Collection("question_category").Where(bson.M{"_id": objectID.(primitive.ObjectID)}).Take(&qc)
//	if err != nil && !sf.MongoNoResult(err) {
//		sf.SLogger().Error(err)
//		return err
//	}
//
//	qc.Categorys = sf.ExamDeal(qc.ExamCategory, qc.ExamChildCategory, qc.Categorys, []string{})
//	qc.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
//	err = sf.DB().Collection("question_category").Save(&qc)
//	if err != nil {
//		sf.SLogger().Error(err)
//		return err
//	}
//
//	return nil
//}
//func (sf *Question) ExamDeal(examCategory string, examChildCategory string, qc []models.QuestionCategoryItem, qks []string) []models.QuestionCategoryItem {
//	type tempData struct {
//		ID struct {
//			QuestionReal int `bson:"question_real"`
//		} `bson:"_id"`
//		Count float64 `bson:"count"`
//	}
//	for i, vo := range qc {
//		var tempQks []string
//		tempQks = append(tempQks, qks...)
//		tempQks = append(tempQks, vo.Title)
//		filter := bson.M{"status": 5}
//
//		for j, questionCategory := range tempQks {
//			filter[fmt.Sprintf("question_category.%d", j)] = questionCategory
//		}
//		filter["exam_category"] = examCategory
//		if examChildCategory != "" {
//			filter["exam_child_category"] = examChildCategory
//		}
//
//		var tempResp []tempData
//		aggregateF := bson.A{bson.M{"$match": filter},
//			bson.M{"$group": bson.M{"_id": bson.M{"question_real": "$question_real"}, "count": bson.M{"$sum": 1}}}}
//		err := sf.DB().Collection("g_interview_questions").Aggregate(aggregateF, &tempResp)
//		if err == nil {
//			for _, ii := range tempResp {
//				tempRespLength := len(tempResp)
//				if ii.ID.QuestionReal == 0 {
//					qc[i].NotRealQuestionCount = int64(ii.Count)
//					if tempRespLength == 1 {
//						qc[i].RealQuestionCount = 0
//					}
//				} else {
//					qc[i].RealQuestionCount = int64(ii.Count)
//					if tempRespLength == 1 {
//						qc[i].NotRealQuestionCount = 0
//					}
//				}
//			}
//			qc[i].QuestionCount = qc[i].NotRealQuestionCount + qc[i].RealQuestionCount
//		} else {
//			qc[i].NotRealQuestionCount = 0
//			qc[i].RealQuestionCount = 0
//			qc[i].QuestionCount = 0
//		}
//
//		if len(vo.ChildCategory) > 0 {
//			qc[i].ChildCategory = sf.ExamDeal(examCategory, examChildCategory, vo.ChildCategory, tempQks)
//		}
//	}
//	return qc
//}

// 试题集合列表
func (sf *Question) ReviewList(c *gin.Context) {
	var param struct {
		Id                string `json:"id"`
		Status            int8   `json:"status"`
		ScoreType         int8   `json:"score_type"`
		Keyword           string `json:"keyword"`
		PageIndex         int64  `json:"page_index"`
		PageSize          int64  `json:"page_size"`
		ExamCategory      string `json:"exam_category"`
		ExamChildCategory string `json:"exam_child_category"`
		ReviewType        int8   `json:"review_type"` // 0 全部 1 班级 2其他
		ClassId           string `json:"class_id"`    // 班级
		ChapterId         string `json:"chapter_id"`  // 小节
		CourseId          string `json:"course_id"`   // 课程ID
	}
	err := c.ShouldBindJSON(&param)
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	filter := bson.M{}
	if param.Keyword != "" {
		filter["$or"] = bson.A{
			bson.M{"manager_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keyword)}}},
			bson.M{"title": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keyword)}}},
			bson.M{"_id": sf.ObjectID(param.Keyword)},
		}
	}
	if param.Status != 0 {
		filter["status"] = param.Status
	}
	if param.ScoreType > 0 {
		filter["score_type"] = param.ScoreType
	}
	if param.Id != "" {
		filter["_id"] = sf.ObjectID(param.Id)
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}

	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.ReviewType != 0 {
		switch param.ReviewType {
		case 1:
			filter["class.class_id"] = bson.M{"$ne": ""}
		case 2:
			filter["class.class_id"] = bson.M{"$eq": ""}
		}
	}
	if param.ClassId != "" {
		filter["class.class_id"] = param.ClassId
	}
	if param.CourseId != "" {
		filter["course.course_id"] = param.CourseId
	}
	if param.ChapterId != "" {
		filter["course.chapter_id"] = param.ChapterId
	}
	uid := c.GetString("user_id")
	managerName := c.GetString("manager_name")
	if !sf.IsAdminManager(uid) {
		filter["manager_name"] = managerName
	}
	resp, total, err := services.NewQuestionSet().List(filter, offset, limit, param.Id)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = resp
	resultInfo["count"] = total
	sf.Success(resultInfo, c)
}

func setList(list []string) []string {
	if len(list) == 0 {
		return list
	}
	setMap := make(map[string]struct{}, 0)
	for _, s := range list {
		_, ok := setMap[s]
		if !ok {
			setMap[s] = struct{}{}
		}
	}
	result := make([]string, 0)
	for k, _ := range setMap {
		result = append(result, k)
	}
	return result
}

// 试题集合编辑
func (sf *Question) ReviewEdit(c *gin.Context) {
	var param struct {
		Id                string   `json:"id"`
		Status            int8     `json:"status"`
		ScoreType         int      `json:"score_type"`
		ExamCategory      string   `json:"exam_category"`       //考试分类
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		Title             string   `json:"title"`
		QuestionIds       []string `json:"question_ids"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	mName, _ := url.QueryUnescape(c.GetHeader("x-user-name"))
	qs, err := services.NewQuestionSet().Edit(param.Id, param.Title, uid, mName, param.Status, param.ScoreType, param.QuestionIds, param.ExamCategory, param.ExamChildCategory)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	sf.Success(qs.Id.Hex(), c)
}

// 试题集合编辑-关联班级
func (sf *Question) ReviewEditClass(c *gin.Context) {
	var param struct {
		Id              string `json:"id"`
		ClassId         string `json:"class_id"`
		ClassName       string `json:"class_name"`
		ClassReviewDate string `json:"class_review_date"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	err = services.NewQuestionSet().EditClass(param.Id, param.ClassId, param.ClassName, param.ClassReviewDate)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(param.Id, c)
}

// 试题集合编辑-关联网课课程
func (sf *Question) ReviewEditCourse(c *gin.Context) {
	var param struct {
		Id                string `json:"id"`
		CourseId          string `json:"course_id"`
		CourseName        string `json:"course_name"`
		ChapterId         string `json:"chapter_id"`
		ChapterName       string `json:"chapter_name"`
		CourseWorkBtnName string `json:"course_work_btn_name"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	err = services.NewQuestionSet().EditCourse(param.Id, param.CourseId, param.CourseName, param.ChapterId, param.ChapterName, param.CourseWorkBtnName)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success("success", c)
}

func (sf *Question) QuestionIsRepeat(c *gin.Context) {
	var param struct {
		QuestionId string `json:"question_id"`
		Name       string `json:"name"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	existId := new(services.Question).GetSimilarStrQuestion("", param.Name)
	returnMsg := ""
	returnData := make(map[string]interface{})
	returnData["repeat_question_ids"] = []string{}
	if len(existId) > 0 {
		returnData["repeat_question_ids"] = existId
		if param.QuestionId == "" {
			returnMsg = fmt.Sprintf("已有相同试题，请勿重复上传！")
		} else {
			if common.InArr(param.QuestionId, existId) {
				returnData["repeat_question_ids"] = []string{}
			} else {
				returnMsg = fmt.Sprintf("已有相同试题，请勿重复保存！")
			}
		}
	}
	sf.Success(returnData, c, returnMsg)
}

// QuestionInfos 查看试题详情（批量）
func (sf *Question) QuestionInfos(c *gin.Context) {
	var param struct {
		QuestionIds []string `json:"question_ids"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil || len(param.QuestionIds) == 0 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var questions []models.GQuestion
	filter := bson.M{"_id": bson.M{"$in": sf.ObjectIDs(param.QuestionIds)}}
	err = sf.DB().Collection("g_interview_questions").Where(filter).Sort("-updated_time").Find(&questions)
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

	returnQuestions := make([]models.GQuestion, 0)
	for _, question := range questions {
		if question.ManagerID != "" {
			question.ManagerName = new(models.Manager).GetManagerName(question.ManagerID)
		}
		// 预生成的答题思路和得分点
		cacheKey := fmt.Sprintf("%s%s", rediskey.GPTQuestionPreview, question.Id.Hex())
		str, err := helper.RedisGet(cacheKey)
		previewItem := make(map[string][]string)
		if str != "" {
			err = json.Unmarshal([]byte(str), &previewItem)
			if err != nil {
				sf.SLogger().Error(err)
			}
		}
		previewIdeas := make([]string, 0)
		previewStandardAnswer := make([]string, 0)
		if _, ok := previewItem["1"]; ok {
			previewIdeas = previewItem["1"]
		}
		if _, ok := previewItem["3"]; ok {
			previewStandardAnswer = previewItem["3"]
		}
		question.PreviewIdeas = previewIdeas
		question.PreviewStandardAnswer = previewStandardAnswer

		returnQuestions = append(returnQuestions, question)
	}
	sf.Success(returnQuestions, c)
}
