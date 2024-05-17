package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/olivere/elastic/v7"
	"github.com/remeh/sizedwaitgroup"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/ffmt.v1"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/es"
	"interview/grpc/client"
	"interview/helper"
	"interview/models"
	"interview/models/appresp"
	"interview/services"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olahol/melody"

	"github.com/garyburd/redigo/redis"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type InterviewGPT struct {
	controllers.Controller
}

var MelodyClient *melody.Melody

func init() {
	MelodyClient = melody.New()
	MelodyClient.HandleConnect(func(s *melody.Session) {
		token := s.Request.FormValue("token")
		userId := new(services.UserToken).Token2User(token, 1)
		if userId == "" {
			s.Write([]byte(`{"code":-1, "data":{},"message":"token异常"}`))
			s.Close()
		}
		s.Set("guid", userId)
	})
	MelodyClient.HandleMessage(new(InterviewGPT).melodyHandleMessage)
	MelodyClient.HandleClose(func(s1 *melody.Session, i int, s2 string) error {
		log.Println("scoket连接断开", i, s2)
		return nil
	})
}

// 用户信息
func (sf *InterviewGPT) GetUserInfo(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo([]string{uid}, appCode, 1)
	sf.Success(userMap[uid], c)

}

// SaveInterviewQuestion 保存试题
//func (sf *InterviewGPT) SaveInterviewQuestion(c *gin.Context) {
//	var param struct {
//		QuestionId     string             `json:"question_id"` // 试题ID
//		Tags           []string           `json:"tags"`
//		Name           string             `json:"name" binding:"required"` // 试题名称
//		Desc           string             `json:"desc"`
//		Answer         string             `json:"answer"`
//		Status         int32              `json:"status"`      // 试题状态
//		CategoryId     string             `json:"category_id"` // 试题分类
//		GPTAnswer      []models.GPTAnswer `json:"gpt_answer"`
//		Year           int                `json:"year"`  // 年份
//		Month          int                `json:"month"` // 月份
//		Day            int                `json:"day"`   // 日
//		Province       string             `json:"province"`
//		City           string             `json:"city"`
//		District       string             `json:"district"`
//		JobTag         string             `json:"job_tag"`         // 岗位标签，如海关、税务局等
//		QuestionSource string             `json:"question_source"` // 试题来源
//	}
//	err := c.ShouldBindJSON(&param)
//	if err != nil {
//		sf.SLogger().Error(err)
//		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
//		return
//	}
//	uid := c.GetHeader("X-XTJ-UID")
//	var question models.GQuestion
//	rdb := sf.RDBPool().Get()
//	defer rdb.Close()
//	// 如果存在试题ID，代表是修改操作
//	if param.QuestionId != "" {
//		questionFilter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
//		err = sf.DB().Collection("g_interview_questions").Where(questionFilter).Take(&question)
//		if err != nil {
//			if sf.MongoNoResult(err) {
//				sf.Error(common.CodeServerBusy, c, "修改失败,试题不存在!")
//				return
//			} else {
//				sf.SLogger().Error(err)
//				sf.Error(common.CodeServerBusy, c)
//				return
//			}
//		}
//		question.Name = param.Name
//		question.Answer = param.Answer
//		question.Status = param.Status
//		question.Desc = param.Desc
//		question.Tags = param.Tags
//		question.CategoryId = param.CategoryId
//		question.Province = param.Province
//		question.City = param.City
//		question.District = param.District
//		question.Year = param.Year
//		question.Month = param.Month
//		question.Day = param.Day
//		question.JobTag = param.JobTag
//		question.QuestionSource = param.QuestionSource
//		question.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
//		err = sf.DB().Collection("g_interview_questions").Save(&question)
//		if err != nil {
//			sf.SLogger().Error(err)
//			sf.Error(common.CodeServerBusy, c)
//			return
//		}
//	} else {
//		// 新增试题
//		question.Name = param.Name
//		question.Answer = param.Answer
//		question.Status = param.Status
//		question.Desc = param.Desc
//		question.Tags = param.Tags
//		question.CreatorUserId = uid
//		question.CategoryId = param.CategoryId
//		question.GPTAnswer = param.GPTAnswer
//		question.Province = param.Province
//		question.City = param.City
//		question.District = param.District
//		question.Year = param.Year
//		question.Month = param.Month
//		question.Day = param.Day
//		question.JobTag = param.JobTag
//		question.QuestionSource = param.QuestionSource
//		_, err = sf.DB().Collection("g_interview_questions").Create(&question)
//		if err != nil {
//			sf.SLogger().Error(err)
//			sf.Error(common.CodeServerBusy, c)
//			return
//		}
//		rdb.Do("RPUSH", rediskey.GPTQuestionAnswerWaiting, question.Id.Hex())
//	}
//
//	rdb.Do("HSET", rediskey.InterviewGPTQuestionId2Name, question.Id.Hex(), question.Name)
//	sf.Success(nil, c)
//}

// DelInterviewQuestion 改变试题状态
func (sf *InterviewGPT) ChangeInterviewQuestionStatus(c *gin.Context) {
	var err error
	var param struct {
		QuestionId string `json:"question_id" binding:"required" msg:"invalid question_id"` // 试题ID
		Status     int32  `json:"status"`                                                   // 试题状态
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, sf.GetValidMsg(err, &param))
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var question models.GQuestion

	filter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
	err = sf.DB().Collection("g_interview_questions").Where(filter).Take(&question)
	if err == nil {
		if param.Status == 9 {
			//删除只有创建者可以删除
			if question.CreatorUserId != uid {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "删除失败!非试题创建者!")
				return
			}
		} else if param.Status == 0 || param.Status == 5 {
			//上下架 只能创建者或者老师
			appCode := c.GetString("APP-CODE")
			userMap := new(models.InterviewGPT).GetUsersInfo([]string{uid}, appCode, 1)
			if userMap[uid].IdentityType == 2 {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c, "操作失败!身份权限不足")
				return
			}
		}
		_, err := sf.DB().Collection("g_interview_questions").Where(filter).Update(map[string]interface{}{"status": param.Status, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "保存失败!")
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
			_, err = sf.DB().Collection("paper").BulkWrite(bulkPullSlice)
			if err != nil {
				sf.SLogger().Error(err)
			}
			_, err = sf.DB().Collection("paper").BulkWrite(bulkSubSlice)
			if err != nil {
				sf.SLogger().Error(err)
			}
		}
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "保存失败!!试题不存在")
		return

	}
	sf.Success(nil, c)
}

// GetInterviewQuestions 查看所有试题
func (sf *InterviewGPT) GetInterviewQuestions(c *gin.Context) {
	var param struct {
		ExamCategory        string   `json:"exam_category"`
		ExamChildCategory   string   `json:"exam_child_category"`
		QuestionCategory    []string `json:"question_category"`
		PageIndex           int64    `json:"page_index"`
		PageSize            int64    `json:"page_size"`
		Keywords            string   `json:"keywords"`
		Province            string   `json:"province"`
		City                string   `json:"city"`
		District            string   `json:"district"`
		JobTag              string   `json:"job_tag"`               // 岗位标签
		QuestionReal        int8     `json:"question_real"`         // 是否真题，0为查全部,1真题，2模拟题
		QuestionContentType int8     `json:"question_content_type"` // 试题类别，0普通题，1漫画题, -1所有
		QuestionIds         string   `json:"question_ids"`          // 试题ID 逗号分隔
		ExerciseStatus      int      `json:"exercise_status"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	ss, _ := json.Marshal(param)
	fmt.Println("question list param:", string(ss))
	// 如果获取到uid，那么就保存用户选择的考试分类和地区
	uid := c.GetHeader("X-XTJ-UID")
	if uid != "" && param.ExamCategory != "" {
		rdb := sf.RDBPool().Get()
		defer rdb.Close()
		keyName := fmt.Sprintf("%s%s", rediskey.UserChoice, uid)
		fieldsValues := []interface{}{keyName}
		province := "不限地区"
		if param.Province != "" {
			province = param.Province
		}
		fieldsValues = append(fieldsValues,
			"exam_category", param.ExamCategory, "exam_child_category", param.ExamChildCategory,
			"province", province, "city", param.City, "district", param.District, "job_tag", param.JobTag)
		_, err = rdb.Do("HMSET", fieldsValues...)
		if err != nil {
			sf.SLogger().Error(err)
		}
	}
	filter := bson.M{"status": 5}
	if param.QuestionContentType != -1 {
		filter["question_content_type"] = param.QuestionContentType
	}
	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"tags": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}}, bson.M{"_id": sf.ObjectID(param.Keywords)},
			bson.M{"question_category": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"gpt_answer.content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.QuestionCategory) > 0 {
		for i, v := range param.QuestionCategory {
			filter[fmt.Sprintf("question_category.%d", i)] = v
		}
	}
	if param.Province != "" {
		filter["province"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Province)}}
	}
	if param.City != "" {
		filter["city"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.City)}}
	}
	if param.District != "" {
		filter["district"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.District)}}
	}
	if param.JobTag != "" {
		filter["job_tag"] = param.JobTag
		delete(filter, "province")
		delete(filter, "city")
		delete(filter, "district")
	}
	if param.QuestionReal != 0 {
		if param.QuestionReal == 1 {
			filter["question_real"] = 1
		} else if param.QuestionReal == 2 {
			filter["question_real"] = 0
		}
	}
	if param.QuestionIds != "" {
		filter["_id"] = bson.M{"$in": sf.ObjectIDs(strings.Split(param.QuestionIds, ","))}
	}
	sortCondition := []string{"-year", "-month", "-day", "+moment", "+objectID"}
	// 热点习题
	if param.QuestionReal == 2 {
		sortCondition = []string{"-created_time", "+objectID"}
	}
	if len(param.QuestionCategory) > 0 && param.QuestionCategory[0] == "热点习题" {
		delete(filter, "question_real")
		delete(filter, "province")
		delete(filter, "city")
		delete(filter, "district")
		delete(filter, "job_tag")
		filter["exam_category"] = "公务员"
		sortCondition = []string{"-created_time", "+objectID"}
	}
	var questions = []models.GQuestion{}
	answerCountMap := make(map[string]int64)
	if uid != "" {
		// 添加做题次数
		f := bson.M{}
		for k, v := range filter {
			if k == "_id" || k == "status" || k == "question_real" {
				continue
			}
			f[k] = v
		}
		f["log_type"] = 1
		f["user_id"] = uid
		var list []models.GAnswerLog
		err = sf.DB().Collection("g_interview_answer_logs").Where(f).Find(&list)
		for _, v := range list {
			if _, ok := answerCountMap[v.QuestionId]; ok {
				answerCountMap[v.QuestionId]++
			} else {
				answerCountMap[v.QuestionId] = 1
			}
		}
	}
	resultInfo := make(map[string]interface{})
	realCount := int64(0)
	notRealCount := int64(0)
	// 已做和未做
	ids := make([]primitive.ObjectID, 0)
	if uid != "" && param.ExerciseStatus == 1 {
		if len(answerCountMap) == 0 {
			resultInfo["list"] = questions
			resultInfo["count"] = 0
			resultInfo["real_count"] = realCount
			resultInfo["not_real_count"] = notRealCount
			sf.Success(resultInfo, c)
			return
		}
		for id := range answerCountMap {
			ids = append(ids, sf.ObjectID(id))
		}
		filter["_id"] = bson.M{"$in": ids}
	} else if uid != "" && param.ExerciseStatus == 2 {
		// 未做
		for id := range answerCountMap {
			ids = append(ids, sf.ObjectID(id))
		}
		if len(ids) > 0 {
			filter["_id"] = bson.M{"$nin": ids}
		}
	}
	fmt.Println(filter)
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_interview_questions").Where(filter).Sort(sortCondition...).Skip(offset).Limit(limit).Find(&questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	tempQuestions := []models.GQuestion{}
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02 15:04:05")
	for _, q := range questions {
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
		if count, ok := answerCountMap[q.Id.Hex()]; ok {
			q.MyAnswerCount = count
		}
		if q.CreatedTime > threeDaysAgo {
			q.IsNew = true
		}
		tempQuestions = append(tempQuestions, q)
	}

	totalCount, _ := sf.DB().Collection("g_interview_questions").Where(filter).Count()

	switch param.QuestionReal {
	case 0:
		filter["question_real"] = 1
		realCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
		filter["question_real"] = 0
		notRealCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	case 1:
		realCount = totalCount
		filter["question_real"] = 0
		notRealCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	case 2:
		notRealCount = totalCount
		filter["question_real"] = 1
		realCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	}

	// 排序
	if param.QuestionIds != "" {
		temp := make([]models.GQuestion, 0)
		for _, qid := range strings.Split(param.QuestionIds, ",") {
			for _, question := range tempQuestions {
				if qid == question.Id.Hex() {
					temp = append(temp, question)
					break
				}
			}
		}
		tempQuestions = temp
	}

	resultInfo["list"] = tempQuestions
	resultInfo["count"] = totalCount
	resultInfo["real_count"] = realCount
	resultInfo["not_real_count"] = notRealCount
	sf.Success(resultInfo, c)

}

// GetInterviewQuestionsWithES 查看所有试题(es)
func (sf *InterviewGPT) GetInterviewQuestionsWithES(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
		Keywords          string   `json:"keywords"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		JobTag            string   `json:"job_tag"`       // 岗位标签
		QuestionReal      int8     `json:"question_real"` // 是否真题，0为查全部,1真题，2模拟题
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 如果获取到uid，那么就保存用户选择的考试分类和地区
	uid := c.GetHeader("X-XTJ-UID")
	if uid != "" {
		rdb := sf.RDBPool().Get()
		defer rdb.Close()
		keyName := fmt.Sprintf("%s%s", rediskey.UserChoice, uid)
		fieldsValues := []interface{}{keyName}
		province := "不限地区"
		if param.Province != "" {
			province = param.Province
		}
		fieldsValues = append(fieldsValues,
			"exam_category", param.ExamCategory, "exam_child_category", param.ExamChildCategory,
			"province", province, "city", param.City, "district", param.District, "job_tag", param.JobTag)
		_, err = rdb.Do("HMSET", fieldsValues...)
		if err != nil {
			sf.SLogger().Error(err)
		}
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
	matchBoost := 2.0
	param.Keywords = strings.ReplaceAll(param.Keywords, "的", "")
	//param.Keywords = strings.ReplaceAll(param.Keywords, "了", "")
	query := elastic.NewBoolQuery()
	// 过滤条件，不参与score计算
	filterList := []elastic.Query{elastic.NewTermQuery("status", 5)}
	if param.ExamCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_category", param.ExamCategory))
	}
	if param.ExamChildCategory != "" {
		filterList = append(filterList, elastic.NewTermQuery("exam_child_category", param.ExamChildCategory))
	}

	for _, t := range param.QuestionCategory {
		filterList = append(filterList, elastic.NewTermQuery("question_category", t))
	}
	if param.Province != "" {
		filterList = append(filterList, elastic.NewWildcardQuery("province", param.Province+"*"))
	}
	if param.City != "" {
		filterList = append(filterList, elastic.NewTermQuery("city", param.City))
	}
	if param.District != "" {
		filterList = append(filterList, elastic.NewTermQuery("district", param.District))
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
	query.Filter(filterList...)
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

	highLightFilter := elastic.NewHighlight()
	highLightFilter = highLightFilter.Fields(elastic.NewHighlighterField("name"), elastic.NewHighlighterField("id"), elastic.NewHighlighterField("desc"), elastic.NewHighlighterField("name_desc"), elastic.NewHighlighterField("name_struct.content.text"))
	highLightFilter.HighlightFilter(true)
	highLightFilter.RequireFieldMatch(true)
	highLightFilter = highLightFilter.PreTags("<font color='blue'>").PostTags("</font>")

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
	ffmt.Puts(query.Source())
	searchResult, err := ElasticClient.Search().
		Index(searchIndex).
		Query(query).
		Highlight(highLightFilter).
		SortBy(
			elastic.NewFieldSort("_score").Desc(),
			elastic.NewFieldSort("year").Desc(),
			elastic.NewFieldSort("month").Desc(),
			elastic.NewFieldSort("day").Desc(),
			//elastic.NewFieldSort("moment").Asc(),
			elastic.NewFieldSort("id").Desc(),
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
	total := searchResult.TotalHits()
	returnQuestionList := make([]models.GQuestion, 0)

	if total > 0 {
		for _, v := range searchResult.Hits.Hits {
			var gQuestion models.GQuestion
			err = json.Unmarshal(v.Source, &gQuestion)
			if err != nil {
				sf.SLogger().Error(err)
			}
			returnQuestionList = append(returnQuestionList, gQuestion)
		}
		questions = returnQuestionList
	}

	tempQuestions := make([]models.GQuestion, 0)
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
				tempCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 1, "user_id": uid, "question_id": questionId}).Count()
				lock.Lock()
				answerCountMap[questionIndex] = tempCount
				lock.Unlock()
			}(i, qId, &sw)
		}
		sw.Wait()
	}
	threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02 15:04:05")
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
		if q.CreatedTime > threeDaysAgo {
			q.IsNew = true
		}
		tempQuestions = append(tempQuestions, q)
	}

	//totalCount, _ := sf.DB().Collection("g_interview_questions").Where(filter).Count()
	//realCount := int64(0)
	//notRealCount := int64(0)
	//switch param.QuestionReal {
	//case 0:
	//	filter["question_real"] = 1
	//	realCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	//	filter["question_real"] = 0
	//	notRealCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	//case 1:
	//	realCount = totalCount
	//	filter["question_real"] = 0
	//	notRealCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	//case 2:
	//	notRealCount = totalCount
	//	filter["question_real"] = 1
	//	realCount, _ = sf.DB().Collection("g_interview_questions").Where(filter).Count()
	//}

	resultInfo := make(map[string]interface{})
	resultInfo["list"] = tempQuestions
	resultInfo["count"] = searchResult.TotalHits()
	//resultInfo["real_count"] = realCount
	//resultInfo["not_real_count"] = notRealCount
	sf.Success(resultInfo, c)

}

// GetTotalInterviewQuestions 查看所有试题（不分页）
func (sf *InterviewGPT) GetTotalInterviewQuestions(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		Keywords          string   `json:"keywords"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	filter := bson.M{"status": 5}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.QuestionCategory) > 0 {
		for i, v := range param.QuestionCategory {
			filter[fmt.Sprintf("question_category.%d", i)] = v
		}
	}
	if param.Keywords != "" {
		filter["$or"] = bson.A{bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"tags": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}}, bson.M{"_id": param.Keywords},
			bson.M{"question_category": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
			bson.M{"gpt_answer.content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Keywords)}}},
		}
	}

	totalCount, _ := sf.DB().Collection("g_interview_questions").Where(filter).Count()
	var questions = []models.GQuestion{}
	err = sf.DB().Collection("g_interview_questions").Where(filter).Sort("-updated_time").Find(&questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = questions
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)
}

// GetInterviewQuestion 查看试题详情
func (sf *InterviewGPT) GetInterviewQuestion(c *gin.Context) {
	QuestionID := c.Query("question_id")
	if QuestionID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	clickCate := c.Query("click_question_category")
	jobTag := c.Query("job_tag")
	provence := c.Query("provence")
	clickCateArr := strings.Split(clickCate, ",")
	// 如果接收到的是短id，转为长id
	if len(QuestionID) <= 8 {
		longID, err := sf.TransferIDLength(QuestionID)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		QuestionID = longID
	}
	var question models.GQuestion
	filter := bson.M{"_id": sf.ObjectID(QuestionID)}
	err := sf.DB().Collection("g_interview_questions").Where(filter).Take(&question)
	appCode := c.GetString("APP-CODE")
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
	year := ""
	month := ""
	day := ""
	if question.Year != 0 {
		year = fmt.Sprintf("%d年", question.Year)
	}
	if question.Month != 0 {
		month = fmt.Sprintf("%d月", question.Month)
	}
	if question.Day != 0 {
		day = fmt.Sprintf("%d日", question.Day)
	}
	question.Date = fmt.Sprintf("%s%s%s%s", year, month, day, question.Moment)

	uid := c.GetHeader("X-XTJ-UID")
	var wg sync.WaitGroup
	var keypointVideoList []models.VideoKeypoints
	var questionVideoList []models.VideoQuestions
	record := make(map[string]string)
	questions := make([]models.GQuestion, 0)
	userMap := make(map[string]models.GUser)

	a := time.Now()
	// 查询用户信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		userMap = new(models.InterviewGPT).GetUsersInfo([]string{question.CreatorUserId}, appCode, 1)
	}()
	// 查询试题视频
	wg.Add(1)
	go func() {
		defer wg.Done()
		questionVideoList = new(models.VideoQuestions).GetQuestionVideoList(QuestionID)
	}()
	// 查询试题视频,知识点视频
	if len(question.QuestionCategory) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keypointVideoList = new(models.VideoKeypoints).GetKeypointVideoList(question.ExamCategory, question.ExamChildCategory, question.QuestionCategory)
		}()
	}
	// 是否有新题 与新题 index
	if uid != "" {
		// 查询 redis 用户足迹
		wg.Add(1)
		go func() {
			defer wg.Done()
			exam := question.ExamCategory
			if question.ExamChildCategory != "" {
				exam += "_" + question.ExamChildCategory
			}
			k := fmt.Sprintf(string(rediskey.UserAnswerKeyPointRecord), uid, exam, jobTag, provence, question.QuestionReal)
			record, _ = helper.RedisHGetAll(k)
		}()

		// 查询当前知识点的 question 列表
		wg.Add(1)
		go func() {
			defer wg.Done()
			f := bson.M{"exam_category": question.ExamCategory}
			if question.ExamChildCategory != "" {
				f["exam_child_category"] = question.ExamChildCategory
			}
			f["status"] = 5
			f["question_real"] = question.QuestionReal
			if clickCate != "" && clickCate != "全部" {
				for i, v := range clickCateArr {
					k := fmt.Sprintf("question_category.%d", i)
					f[k] = v
				}
			}
			s1 := time.Now()
			_ = sf.DB().Collection("g_interview_questions").Where(f).Sort([]string{"-year", "-month", "-day", "-created_time"}...).Fields(bson.M{"_id": 1, "created_time": 1}).Find(&questions)
			fmt.Println("question_info find cost:", time.Since(s1).Milliseconds())
		}()
	}
	wg.Wait()
	fmt.Println("question_info sync all cost:", time.Since(a).Milliseconds())
	if _, ok := userMap[question.CreatorUserId]; ok {
		question.UserName = userMap[question.CreatorUserId].Nickname
	}
	res := common.StructToMap(&question)
	res["id"] = question.Id.Hex()
	res["category_video_list"] = keypointVideoList
	res["question_video_list"] = questionVideoList
	res["has_new"] = false
	res["new_question_id"] = ""
	res["new_question_id_index"] = 0
	// 是否有新题
	res["has_new"], res["new_question_id"], res["new_question_id_index"] = services.DelHasNew(record, questions, clickCateArr, QuestionID)
	sf.Success(res, c)
}

// TransferIDLength 长短ID转换
func (sf *InterviewGPT) TransferIDLength(questionID string) (ID string, err error) {
	if questionID == "" {
		return "", errors.New("试题ID不能为空")
	}
	var t models.GIDMap
	filter := bson.M{}
	if len(questionID) == 32 || len(questionID) == 24 {
		filter["long_id"] = questionID
		err := sf.DB().Collection("g_id_map").Where(filter).Take(&t)
		if err != nil {
			if sf.MongoNoResult(err) {
				// 创建
				dataCount, _ := sf.DB().Collection("g_id_map").Count()
				t.ShortID = strconv.Itoa(int(dataCount))
				t.LongID = questionID
				_, err = sf.DB().Collection("g_id_map").Create(&t)
				if err != nil {
					dataCount, _ = sf.DB().Collection("g_id_map").Count()
					t.ShortID = strconv.Itoa(int(dataCount))
					_, err = sf.DB().Collection("g_id_map").Create(&t)
					if err != nil {
						return "", err
					} else {
						return t.ShortID, nil
					}
				} else {
					return t.ShortID, nil
				}

			} else {
				return "", err
			}
		} else {
			return t.ShortID, nil
		}
	} else if len(questionID) <= 8 {
		// 短id查询长id，不存在时无需创建
		filter["short_id"] = questionID
		err := sf.DB().Collection("g_id_map").Where(filter).Take(&t)
		if err != nil {
			return "", err
		} else {
			return t.LongID, nil
		}
	} else {
		sf.SLogger().Error("试题id是：", questionID)
		return "", errors.New("试题ID长度有误")
	}
}

// SaveAnswerLog 保存答案记录
func (sf *InterviewGPT) SaveAnswerLog(c *gin.Context) {
	var param struct {
		ReviewLogId  string  `json:"review_log_id"`
		ReviewId     string  `json:"review_id"`
		QuestionId   string  `json:"question_id" binding:"required"`
		VoiceUrl     string  ` json:"voice_url"`    //语音url
		VoiceLength  float64 `json:"voice_length"`  //语音时长
		VoiceText    string  `json:"voice_text"`    //语音文本
		SID          string  `json:"sid"`           // 科大讯飞SID
		PracticeType int8    `json:"practice_type"` // 练习种类，11是看题-普通模式，12是看题-对镜模式，13是看题-考官模式 21是听题-普通模式，22是听题-对镜模式，23是听题-考官模式
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.PracticeType == 0 {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c, "未传练习模式参数")
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var answerLog models.GAnswerLog
	var question models.GQuestion
	err = sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(param.QuestionId)}).Take(&question)
	if err == nil {
		answerLog.ExamCategory = question.ExamCategory
		answerLog.ExamChildCategory = question.ExamChildCategory
		answerLog.QuestionCategory = question.QuestionCategory
		answerLog.Province = question.Province
		answerLog.City = question.City
		answerLog.District = question.District
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "无效的试题")
		return
	}
	// 查询是否已经提交答案
	answerLog.SID = param.SID
	answerLog.QuestionName = question.Name
	answerLog.NameStruct = question.NameStruct
	answerLog.QuestionContentType = question.QuestionContentType
	answerLog.UserId = uid
	answerLog.QuestionId = param.QuestionId
	answerLog.Answer = []models.GAnswer{{VoiceUrl: param.VoiceUrl, VoiceContent: []models.VoiceContent{}, VoiceLength: sf.TransitionFloat64(param.VoiceLength, -1), VoiceText: param.VoiceText}}
	answerLog.PracticeType = param.PracticeType
	if param.ReviewId != "" && param.ReviewLogId == "" {
		reviewlog := new(models.ReviewLog)
		err = sf.DB().Collection("review_log").Where(bson.M{"review_id": param.ReviewId, "user_id": uid}).Take(reviewlog)
		if err != nil {
			if sf.MongoNoResult(err) {
				reviewlog, _ = services.NewQuestionSet().MakeLog(uid, param.ReviewId, 0, "")
			} else {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
		}
		param.ReviewLogId = reviewlog.Id.Hex()
	}
	answerLog.ReviewLogId = param.ReviewLogId
	answerLog.ReviewId = param.ReviewId
	answerLog.LogType = 1
	_, err = sf.DB().Collection("g_interview_answer_logs").Create(&answerLog)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	tempEsClient, err := es.CreateEsClient()
	if err == nil {
		err = es.AnswerLogAddToEs(tempEsClient, context.Background(), []models.GAnswerLog{answerLog})
		if err != nil {
			sf.SLogger().Error(err)
		}
	} else {
		sf.SLogger().Error(err)
	}
	rdb := sf.RDBPool().Get()
	defer rdb.Close()

	var needGPTComment = true
	if answerLog.ReviewId != "" {
		review := new(models.Review)
		sf.DB().Collection("review").Where(bson.M{"_id": sf.ObjectID(answerLog.ReviewId)}).Take(review)
		if review.ScoreType == 2 {
			needGPTComment = false
		}
	}
	// 面试云的log不参与点评
	if answerLog.LogType == 2 {
		needGPTComment = false
	}
	if needGPTComment {
		rdb.Do("RPUSH", rediskey.GPTCommentWaiting, answerLog.Id.Hex())
	}

	// 如果有测评
	if param.ReviewLogId != "" {
		err = services.NewQuestionSet().UpdateReviewLog(param.ReviewLogId, answerLog)
		if err != nil {
			sf.SLogger().Error(err)
		}
	}
	// 异步任务
	if answerLog.LogType == 1 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					sf.SLogger().Error("deal user keypoint err:", r)
				}
			}()
			err := new(models.UserAnswer).CreateLog(answerLog, question)
			if err != nil {
				sf.SLogger().Error("create user answer err:", err)
			}
		}()
	}
	sf.Success(map[string]interface{}{"log_id": answerLog.Id.Hex()}, c)
}

// SaveGPTThoughtIndex 保存当前用户查看的是第几个思路
func (sf *InterviewGPT) SaveGPTThoughtIndex(c *gin.Context) {
	var param struct {
		AnswerLogID  string `json:"answer_log_id" bson:"answer_log_id"`
		ThoughtIndex int8   `json:"thought_index" bson:"thought_index"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	uid := c.GetHeader("X-XTJ-UID")
	filter := bson.M{"log_type": 1, "user_id": uid, "_id": sf.ObjectID(param.AnswerLogID)}
	var answerLog models.GAnswerLog
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Take(&answerLog)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "未查到该条记录")
		return
	}

	_, err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Update(bson.M{"thought_index": param.ThoughtIndex})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "更新失败")
		return
	}
	sf.Success(nil, c)
}

type RecMessage struct {
	MsgType int8   `json:"msg_type"`
	ID      string `json:"id"`
}
type RespMessage struct {
	Content string `json:"content"`
	MsgType int8   `json:"msg_type"`
	ID      string `json:"id"`
	End     bool   `json:"end"`
}

func (sf *InterviewGPT) getGPTComment(answerLogId string, times int) string {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	isHav, _ := redis.Bool(rdb.Do("SISMEMBER", rediskey.GPTCommentSuccess, answerLogId))
	if isHav {
		var answerLog models.GAnswerLog
		err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 1, "_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
		if err == nil {
			return answerLog.GPTComment.Content
		} else {
			sf.SLogger().Error(err)
			return ""
		}
	} else {
		time.Sleep(2 * time.Second)
		if times < 100 {
			return sf.getGPTComment(answerLogId, times+1)
		} else {
			return ""
		}
	}
}

func (sf *InterviewGPT) getGPTCustomAnswer(answerLogId string, times int) string {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	isHav, _ := redis.Bool(rdb.Do("SISMEMBER", rediskey.GPTCustomSuccess, answerLogId))
	if isHav {
		var answerLog models.GCustomQuestion
		err := sf.DB().Collection("g_custom_questions").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
		if err == nil {
			return answerLog.GPTAnswer.Content
		} else {
			sf.SLogger().Error(err)
			return ""
		}
	} else {
		time.Sleep(2 * time.Second)
		if times < 100 {
			return sf.getGPTCustomAnswer(answerLogId, times+1)
		} else {
			return ""
		}
	}
}

func (sf *InterviewGPT) melodyHandleMessage(s *melody.Session, msg []byte) {
	rm := RecMessage{}
	err := json.Unmarshal(msg, &rm)
	if err == nil {
		if rm.MsgType == 1 {
			//回传gpt 答题思路
			answerLogId := rm.ID
			var answerLog models.GAnswerLog
			resp := RespMessage{MsgType: rm.MsgType, ID: rm.ID}
			err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"log_type": 1, "_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
			if err == nil {
				if answerLog.GPTComment.Content == "" {
					resp.Content = sf.getGPTComment(answerLogId, 0)
					resp.End = true
					sf.SocketSuccess(resp, s)
				} else {
					//已经生成过了
					sf.SocketSuccess(resp, s)
				}
			} else {
				sf.SLogger().Error(err)
				sf.SocketError(common.InvalidId, s)
			}
		} else if rm.MsgType == 2 {
			//回传学员自主提问中gpt的生成答案
			answerLogId := rm.ID
			var answerLog models.GCustomQuestion
			resp := RespMessage{MsgType: rm.MsgType, ID: rm.ID}
			err := sf.DB().Collection("g_custom_questions").Where(bson.M{"_id": sf.ObjectID(answerLogId)}).Take(&answerLog)
			if err == nil {
				if answerLog.GPTAnswer.Content == "" {
					resp.Content = sf.getGPTCustomAnswer(answerLogId, 0)
					resp.End = true
					sf.SocketSuccess(resp, s)
				} else {
					//已经生成过了
					sf.SocketSuccess(resp, s)
				}
			} else {
				sf.SLogger().Error(err)
				sf.SocketError(common.InvalidId, s)
			}
		} else {
			sf.SocketError(common.CodeServerBusy, s, "无效msg_type")
		}

	} else {
		sf.SocketError(common.CodeInvalidParam, s)
	}
}

func (sf *InterviewGPT) SendGPT(c *gin.Context) {
	MelodyClient.HandleRequest(c.Writer, c.Request)
}

// GetAnswerLogs 查看答题记录question_id
func (sf *InterviewGPT) GetAnswerLogs(c *gin.Context) {
	var param struct {
		TeacherCorrect    int8     `json:"teacher_correct"` // 1表示是老师测评问题的回答列表
		ReviewId          string   `json:"review_id"`       //
		ReviewLogId       string   `json:"review_log_id"`   //
		QuestionId        string   `json:"question_id"`
		AnswerUserId      string   `json:"answer_user_id"`
		PageIndex         int64    `json:"page_index"`
		PageSize          int64    `json:"page_size"`
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		LogType           int8     `json:"log_type"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	var logs []models.GAnswerLog
	filter := bson.M{"user_id": uid}
	if param.LogType != 0 {
		filter["log_type"] = param.LogType
	}
	if param.QuestionId != "" {
		if param.TeacherCorrect == 1 {
			if param.AnswerUserId != "" {
				filter = bson.M{"user_id": param.AnswerUserId}
			}
			if param.ReviewId != "" {
				review := new(models.Review)
				err = sf.DB().Collection("review").Where(bson.M{"_id": sf.ObjectID(param.ReviewId)}).Take(review)
				if review.Class.ClassID != "" {
					classMate := services.NewQuestionSet().ClassMate(review.Class.ClassID)
					if len(classMate) == 0 {
						sf.Error(common.CodeServerBusy, c, "班级未关联学员")
						return
					}
					filter = bson.M{"user_id": bson.M{"$in": common.GetMapKeys(classMate)}}
				}
			}
		} else {
			filter = bson.M{"user_id": uid}
		}
		if param.ReviewId != "" {
			filter["review_id"] = param.ReviewId
		}
		if param.ReviewLogId != "" {
			filter["review_log_id"] = param.ReviewLogId
		}
		// 如果接收到的是短id，转为长id
		if len(param.QuestionId) <= 8 {
			longID, err := sf.TransferIDLength(param.QuestionId)
			if err != nil {
				sf.SLogger().Error(err)
				sf.Error(common.CodeServerBusy, c)
				return
			}
			param.QuestionId = longID
		}
		filter["question_id"] = param.QuestionId
	} else {
		filter = bson.M{"user_id": uid}
		filter["user_id"] = uid
		if param.LogType != 0 {
			filter["log_type"] = param.LogType
		}
		if param.ExamCategory != "" {
			filter["exam_category"] = param.ExamCategory
		}
		if param.ExamChildCategory != "" {
			filter["exam_child_category"] = param.ExamChildCategory
		}
		if len(param.QuestionCategory) > 0 {
			for i, v := range param.QuestionCategory {
				filter[fmt.Sprintf("question_category.%d", i)] = v
			}
		}
	}
	totalCount, _ := sf.DB().Collection("g_interview_answer_logs").Where(filter).Count()
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	err = sf.DB().Collection("g_interview_answer_logs").Where(filter).Sort("-updated_time").Skip(offset).Limit(limit).Find(&logs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	if len(logs) > 0 {
		uids := []string{}
		for _, v := range logs {
			uids = append(uids, v.UserId)
		}
		appCode := c.GetString("APP-CODE")
		userMap := new(models.InterviewGPT).GetUsersInfo(uids, appCode, 1)
		for i, v := range logs {
			logs[i].UserName = userMap[v.UserId].Nickname
			logs[i].Avatar = userMap[v.UserId].Avatar
			logs[i].QuestionName = new(models.GQuestion).GetQuestionName(v.QuestionId)
			if logs[i].QuestionName == "" {
				logs[i].QuestionName = new(models.InterviewQuestion).GetQuestionName(v.QuestionId)

			}
		}

		// 如果是老师点评时，排序 按照未点评 到 已点评
		if param.TeacherCorrect == 1 {
			sort.Slice(logs, func(i, j int) bool {
				if logs[i].CorrectStatus == logs[j].CorrectStatus {
					return logs[i].UpdatedTime > logs[j].UpdatedTime
				} else {
					return logs[i].CorrectStatus < logs[j].CorrectStatus
				}
			})
		}
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = logs
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)

}

func (a *InterviewGPT) RecordAnswerPoint(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	if uid == "" {
		a.Error(common.PermissionDenied, c)
		return
	}
	var param struct {
		QuestionID        string   `json:"question_id" binding:"required"`
		QuestionReal      int8     `json:"question_real"`
		ExamCategory      string   `json:"exam_category" binding:"required"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		JobTag            string   `json:"job_tag"`
		Provence          string   `json:"provence"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		a.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	if param.QuestionID == "" {
		a.Error(common.CodeInvalidParam, c)
		return
	}
	// 查询试题
	var question models.GQuestion
	err = a.DB().Collection("g_interview_questions").Where(bson.M{"_id": a.ObjectID(param.QuestionID)}).Take(&question)
	if err != nil {
		a.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	if len(question.QuestionCategory) == 0 {
		question.QuestionCategory = []string{"全部"}
	}
	services.InitInterviewQuestionService().RecordUserQuestion(uid, param.JobTag, param.Provence, question)
	a.Success(true, c)
}

// GetAnswerLog 查看答案记录详情
func (sf *InterviewGPT) GetAnswerLog(c *gin.Context) {
	logID := c.Query("answer_log_id")
	uid := c.Query("uid")
	if logID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if uid == "" {
		uid = c.GetHeader("X-XTJ-UID")
	}
	var tempLog models.GAnswerLog
	filter := bson.M{"_id": sf.ObjectID(logID), "user_id": uid, "log_type": 1}
	err := sf.DB().Collection("g_interview_answer_logs").Where(filter).Take(&tempLog)
	if err != nil {
		if sf.MongoNoResult(err) {
			sf.Error(common.CodeServerBusy, c, "答案记录不存在")
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	var question models.GQuestion
	err = sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(tempLog.QuestionId)}).Take(&question)
	if err == nil {
		tempLog.GPTAnswer = question.GPTAnswer[0]
		tempLog.GPTAnswers = question.GPTAnswer
	} else {
		sf.SLogger().Error(err)
	}
	appCode := c.GetString("APP-CODE")
	userMap := new(models.InterviewGPT).GetUsersInfo([]string{tempLog.UserId}, appCode, 1)
	tempLog.UserName = userMap[tempLog.UserId].Nickname
	tempLog.Avatar = userMap[tempLog.UserId].Avatar
	tempLog.GPTStandardAnswer = question.Answer

	if tempLog.ReviewId != "" {
		review := new(models.Review)
		sf.DB().Collection("review").Where(bson.M{"_id": sf.ObjectID(tempLog.ReviewId)}).Take(review)
		tempLog.ScoreType = review.ScoreType
		tempLog.ClassId = review.Class.ClassID
	}
	sf.Success(tempLog, c)
}

// GetAnswerLog 删除答题记录
func (sf *InterviewGPT) DelAnswerLog(c *gin.Context) {
	logID := c.Query("answer_log_id")
	if logID == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	filter := bson.M{"_id": sf.ObjectID(logID), "user_id": uid, "log_typ": 1}
	_, err := sf.DB().Collection("g_interview_answer_logs").Where(filter).Update(map[string]interface{}{"is_deleted": 1, "updated_time": time.Now().Format("2006-01-02 15:04:05")})
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "删除失败!")
		return
	}
	sf.Success(nil, c)
}
func (sf *InterviewGPT) SetWebRelayCache(c *gin.Context) {
	var err error
	var param map[string]interface{}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	r, err := json.Marshal(param)
	rand.Seed(time.Now().UnixNano())
	code := int64(rand.Intn(9))
	t := time.Now().Unix() + code
	if err == nil {
		rdb.Do("SET", fmt.Sprintf("%s:%d", rediskey.WebRelayCache, t), r)
	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(t, c)

}

func (sf *InterviewGPT) GetWebRelayCache(c *gin.Context) {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	ID := c.Query("id")
	r, err := redis.String(rdb.Do("GET", fmt.Sprintf("%s:%s", rediskey.WebRelayCache, ID)))
	if err == nil {
		res := map[string]interface{}{}
		err := json.Unmarshal([]byte(r), &res)
		if err == nil {
			sf.Success(res, c)
			return
		} else {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}

	} else {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

}

// GetInterviewQuestions 查看所有试题
func (sf *InterviewGPT) GetQuestionCategory(c *gin.Context) {
	var err error
	filter := bson.M{"status": 5}
	totalCount, _ := sf.DB().Collection("g_interview_question_category").Where(filter).Count()
	var qc = []models.GQuestionCategory{}
	err = sf.DB().Collection("g_interview_question_category").Where(filter).Sort("+created_time").Find(&qc)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	resultInfo := make(map[string]interface{})
	resultInfo["list"] = qc
	resultInfo["count"] = totalCount
	sf.Success(resultInfo, c)

}

// GetRandomInterviewQuestion 获取随机试题详情
func (sf *InterviewGPT) GetRandomInterviewQuestion(c *gin.Context) {
	var param struct {
		QuestionId        string   `json:"question_id"` //试题ID
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		QuestionCategory  []string `json:"question_category"`
		Province          string   `json:"province"`
		City              string   `json:"city"`
		District          string   `json:"district"`
		QuestionReal      int8     `json:"question_real"` // 是否真题，0为查全部,1真题，2模拟题
		JobTag            string   `json:"job_tag"`       // 岗位标签，如海关、税务局等
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 临时增加
	if param.QuestionId != "" {
		var question models.GQuestion

		filter := bson.M{"_id": sf.ObjectID(param.QuestionId)}
		err = sf.DB().Collection("g_interview_questions").Where(filter).Take(&question)
		year := ""
		month := ""
		day := ""
		if question.Year != 0 {
			year = fmt.Sprintf("%d年", question.Year)
		}
		if question.Month != 0 {
			month = fmt.Sprintf("%d月", question.Month)
		}
		if question.Day != 0 {
			day = fmt.Sprintf("%d日", question.Day)
		}
		question.Date = fmt.Sprintf("%s%s%s%s", year, month, day, question.Moment)
		if question.AreaCodes == nil {
			question.AreaCodes = []string{}
		}
		sf.Success(question, c)
		return
	}

	uid := c.GetHeader("X-XTJ-UID")
	answerLogs := make([]models.GAnswerLog, 0)
	logFilter := bson.M{"user_id": uid, "log_type": 1}
	if param.ExamCategory != "" {
		logFilter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		logFilter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.QuestionCategory) != 0 {
		logFilter["question_category"] = param.QuestionCategory
	}

	err = sf.DB().Collection("g_interview_answer_logs").Where(logFilter).Fields(bson.M{"_id": 1, "question_id": 1}).Find(&answerLogs)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var answerLogIds []primitive.ObjectID
	for _, tempLog := range answerLogs {
		answerLogIds = append(answerLogIds, sf.ObjectID(tempLog.QuestionId))
	}
	filter := bson.M{"status": 5, "question_content_type": 0}
	if len(answerLogIds) > 0 {
		filter["_id"] = bson.M{"$nin": answerLogIds}
	}
	if param.QuestionReal != 0 {
		if param.QuestionReal == 1 {
			filter["question_real"] = 1
		} else if param.QuestionReal == 2 {
			filter["question_real"] = 0
		}
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if len(param.QuestionCategory) != 0 {
		filter["question_category"] = param.QuestionCategory
	}
	if param.Province != "" {
		filter["province"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.Province)}}
	}
	if param.City != "" {
		filter["city"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.City)}}
	}
	if param.District != "" {
		filter["district"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(param.District)}}
	}
	if param.JobTag != "" {
		filter["job_tag"] = param.JobTag
	}
	filter["tts_url"] = bson.M{"$exists": true}
	filter["tts_url.male_voice_url"] = bson.M{"$ne": ""}
	filter["tts_url.female_voice_url"] = bson.M{"$ne": ""}

	questions := make([]models.GQuestion, 0)
	aggregateF := bson.A{bson.M{"$match": filter},
		bson.M{"$sample": bson.M{"size": 1}}}
	err = sf.DB().Collection("g_interview_questions").Aggregate(aggregateF, &questions)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var question models.GQuestion
	if len(questions) >= 1 {
		question = questions[0]
	} else {
		delete(filter, "_id")
		_ = sf.DB().Collection("g_interview_questions").Aggregate(aggregateF, &questions)
		if len(questions) >= 1 {
			question = questions[0]
		} else {
			sf.Error(common.CodeServerBusy, c, "该分类下试题暂不支持听题练习~")
			return
		}
	}
	year := ""
	month := ""
	day := ""
	if question.Year != 0 {
		year = fmt.Sprintf("%d年", question.Year)
	}
	if question.Month != 0 {
		month = fmt.Sprintf("%d月", question.Month)
	}
	if question.Day != 0 {
		day = fmt.Sprintf("%d日", question.Day)
	}
	question.Date = fmt.Sprintf("%s%s%s%s", year, month, day, question.Moment)
	if question.AreaCodes == nil {
		question.AreaCodes = []string{}
	}
	sf.Success(question, c)
}

func (sf *InterviewGPT) GetTTSPrompt(c *gin.Context) {
	var param struct {
		PromptType int8 `json:"prompt_type"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	rdb := sf.RDBPool().Get()
	defer rdb.Close()

	isNeedQueryFromMongo := false
	prompt := appresp.TTSPromptResp{}
	keyName := fmt.Sprintf("%s%d", rediskey.TTSPrompt, param.PromptType)
	res, err := redis.Values(rdb.Do("HGETALL", keyName))
	if err != nil {
		sf.SLogger().Error(err)
		isNeedQueryFromMongo = true
	}
	if len(res) == 0 {
		isNeedQueryFromMongo = true
	}

	err = redis.ScanStruct(res, &prompt)
	if err != nil {
		sf.SLogger().Error(err)
		isNeedQueryFromMongo = true
	}
	if isNeedQueryFromMongo {
		err = sf.DB().Collection("tts_prompt").Where(bson.M{"prompt_type": param.PromptType}).Take(&prompt)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c, "获取prompt失败了")
			return
		}

		values := []interface{}{keyName,
			"male_head_prompt", prompt.MaleHeadPrompt,
			"male_tail_prompt", prompt.MaleTailPrompt,
			"female_head_prompt", prompt.FeMaleHeadPrompt,
			"female_tail_prompt", prompt.FeMaleTailPrompt,
			"prompt_type", prompt.PromptType,
			"head_voice_length", prompt.HeadVoiceLength,
			"tail_voice_length", prompt.TailVoiceLength,
		}
		_, err = rdb.Do("HMSET", values...)
		_, err = rdb.Do("EXPIRE", keyName, 60*60*24)
	}
	sf.Success(prompt, c)
}

func (sf *InterviewGPT) IsShowPage(c *gin.Context) {
	resp, err := helper.RedisGet(string(rediskey.IsShowGPTPage))
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c, "获取是否展示页面失败了")
		return
	}
	isShow := false
	if resp == "true" {
		isShow = true
	}
	sf.Success(map[string]bool{"is_show": isShow}, c)
}

func (sf *InterviewGPT) WechatSign(c *gin.Context) {
	if services.WeixinCheck(c.Query("timestamp"), c.Query("nonce"), c.Query("signature")) {
		c.String(200, c.Query("echostr"))
	} else {
		c.String(200, "fail")
	}
}

func (sf *InterviewGPT) WechatMsg(c *gin.Context) {
	rd, err := c.GetRawData()
	if err != nil {
		sf.SLogger().Error(err)
	}
	sf.SLogger().Info("WechatMsg rawdata:" + string(rd))

	c.String(200, "success")
	return
}

func (sf *InterviewGPT) ActivityNewYear(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "newyear")
	if str == "" {
		str = `[
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/倒计时3天.jpg",
        "20240207"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/倒计时2天.jpg",
        "20240208"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/倒计时1天.jpg",
        "20240209"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/1.初一-面试AI活动.jpg",
        "20240210"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/2.初二-面试AI活动.jpg",
        "20240211"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/3.初三-面试AI活动.jpg",
        "20240212"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/4.初四-面试AI活动.jpg",
        "20240213"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/5.初五-面试AI活动.jpg",
        "20240214"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/6.初六-面试AI活动.jpg",
        "20240215"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/7.初七-面试AI活动.jpg",
        "20240216"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/8.初八-面试AI活动.jpg",
        "20240217"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/9.初九-面试AI活动.jpg",
        "20240218"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/10.初十-面试AI活动.jpg",
        "20240219"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/11.十一-面试AI活动.jpg",
        "20240220"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/12.十二-面试AI活动.jpg",
        "20240221"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/13.十三-面试AI活动.jpg",
        "20240222"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/14.十四-面试AI活动.jpg",
        "20240223"
    ],
    [
        "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/activity/15.十五-面试AI活动.jpg",
        "20240224"
    ]
]`
	}
	nowStr := time.Now().Format("20060102")
	resp := make([][]string, 0)
	resp2 := ""
	json.Unmarshal([]byte(str), &resp)
	for _, stringss := range resp {
		if stringss[1] == nowStr {
			resp2 = stringss[0]
		}
	}
	sf.Success(map[string]any{
		"url": resp2,
	}, c)
}

func (sf *InterviewGPT) ActivityNeiMeng(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "neimeng")
	resp := make(map[string]string)
	json.Unmarshal([]byte(str), &resp)
	//str := map[string]string{
	//	"banner": "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"share":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"popup":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//}
	nowStr := time.Now().Format("20060102")
	if "20240310" < nowStr {
		resp = make(map[string]string)
	}
	sf.Success(resp, c)
}

func (sf *InterviewGPT) Activity0229(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "0229")
	resp := make(map[string]string)
	json.Unmarshal([]byte(str), &resp)
	//str := map[string]string{
	//	"banner": "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"share":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"popup":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//}
	nowStr := time.Now().Format("20060102")
	if len(resp) == 0 || resp["end_date"] < nowStr {
		resp = make(map[string]string)
	}
	sf.Success(resp, c)
}

func (sf *InterviewGPT) Activity022902(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "022902")
	resp := make(map[string]map[string]string)
	resp2 := make(map[string]string)
	json.Unmarshal([]byte(str), &resp)
	//str := map[string]string{
	//	"banner": "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"share":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"popup":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"page_url":  "/page/index/index?question_type=xxx", // 跳转页面的地址
	//	"app_id":  "公务员/税务",
	//  "moment":  "1", // 跳转时机 1、
	//}
	nowStr := time.Now().Format("20060102")
	if _, ok := resp[nowStr]; ok {
		resp2 = resp[nowStr]
	} else if _, ok = resp["default"]; ok {
		resp2 = resp["default"]
		if len(resp2) == 0 || resp2["end_date"] < nowStr {
			resp2 = make(map[string]string)
		}
	}

	sf.Success(resp2, c)
}

func (sf *InterviewGPT) Activity0301(c *gin.Context) {
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "0301")
	resp := make(map[string]string)
	json.Unmarshal([]byte(str), &resp)
	//str := map[string]string{
	//	"banner": "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"share":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//	"popup":  "https://web.xtjstatic.cn/xtj-interview-ai/shareImg.png",
	//}
	nowStr := time.Now().Format("20060102")
	if len(resp) != 0 {
		if "20240303" > nowStr {
			resp["share_moni"] = ""
		}
		if "20240305" > nowStr {
			resp["share_zhenti"] = ""
		}
	}
	if len(resp) == 0 || resp["end_date"] < nowStr {
		resp = make(map[string]string)
	}
	sf.Success(resp, c)
}

// 分享图
//
//	{
//		"公务员": {
//			"福建": {
//				"模拟试卷": {
//					"pages/index/index": "http://bai.com",
//					"question/sameSourceQuestionList/index": "http://bai.com"
//				}
//			}
//		}
//	}
func (sf *InterviewGPT) ShareImg(c *gin.Context) {
	var examCategory = c.Query("exam_category")
	var page = c.Query("page")
	var questionCategory = c.Query("question_category")
	var province = c.Query("province")
	var questionSource = c.Query("question_source") // 试题来源

	var (
		resp2 string
	)
	str, _ := helper.RedisGet(string(rediskey.ActivityPrefix) + "share")
	resp := make(map[string]map[string]map[string]map[string]string)
	json.Unmarshal([]byte(str), &resp)
	if _, ok := resp[examCategory]; ok {
		if _, ok = resp[examCategory][province]; ok {
			if _, ok = resp[examCategory][province][questionCategory]; ok {
				if _, ok = resp[examCategory][province][questionCategory][page+questionSource]; ok {
					resp2 = resp[examCategory][province][questionCategory][page+questionSource]
				}
			}
		}
	}
	sf.Success(map[string]string{
		"url": resp2,
	}, c)
}

func (sf *InterviewGPT) SpeechTextTest(c *gin.Context) {
	audoUrl := c.Query("audo_url")
	taskid, err := client.CreateRecTask(audoUrl, "https://dev-api.xtjzx.cn/interview/app/n/v1/g/speechtext/callback")
	sf.Success(map[string]any{
		"taskId": taskid,
		"err":    err,
	}, c)
}
func (sf *InterviewGPT) SpeechTextCallback(c *gin.Context) {
	//code	int64	任务状态码，0为成功，其他：失败；详见 状态码说明
	//message	string	失败原因文字描述，成功时此值为空
	//requestId	uint64	任务唯一标识，与录音识别请求中返回的 TaskId 一致。数据格式必须设置为 Uint64
	//appid	uint64	腾讯云应用 ID
	//projectid	int64	腾讯云项目 ID
	//audioUrl	string	语音 url，如创建任务时为上传数据的方式，则不包含该字段
	//text	string	识别出的结果文本
	//resultDetail	string	包含 详细识别结果，如创建任务时 ResTextFormat 为0，则不包含该字段
	//audioTime	double	语音总时长
	code := c.PostForm("code")
	message := c.PostForm("message")
	requestId := c.PostForm("requestId")
	//resultDetail := c.PostForm("resultDetail")
	text := c.PostForm("text")
	audioTimeStr := c.PostForm("audioTime")
	var audioTime float64
	if audioTimeStr != "" {
		audioTime, _ = strconv.ParseFloat(audioTimeStr, 64)
		audioTimeStr = strconv.FormatFloat(audioTime, 'f', 1, 64) // 保留一位小数
		audioTime, _ = strconv.ParseFloat(audioTimeStr, 64)
	}
	if strings.TrimSpace(text) == "" {
		text = "-"
	}
	sf.SLogger().Info("SpeechTextCallback code:", code, " message:", message, " requestId:", requestId)
	var glog = new(models.GAnswerLog)
	err := sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"speech_text_task_id": requestId}).Take(glog)
	if err != nil {
		if sf.MongoNoResult(err) {
			var comment = new(models.AnswerComment)
			err = sf.DB().Collection("interview_comment_logs").Where(bson.M{"speech_text_task_id": requestId}).Take(comment)
			if err == nil {
				comment.Comment.VoiceText = text
				comment.Comment.VoiceLength = audioTime
				comment.Comment.Status = 1
				_, err = sf.DB().Collection("interview_comment_logs").Where(bson.M{"_id": comment.Id}).Update(comment)
			} else {
				sf.SLogger().Error("SpeechTextCallback not found requestId: ", requestId, " err: ", err, " text: ", text)
			}
		} else {
			sf.SLogger().Error("SpeechTextCallback err:", err)
		}
	} else {
		glog.Answer[0].VoiceLength = audioTime
		glog.Answer[0].VoiceText = text
		glog.Answer[0].Status = 1
		_, err = sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"_id": glog.Id}).Update(glog)
		if err != nil {
			sf.SLogger().Error("SpeechTextCallback Update err:", err, " logId:", glog.Id.Hex())
		}
	}

	// { "code" : 0, "message" : "success" }
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
	return
}
