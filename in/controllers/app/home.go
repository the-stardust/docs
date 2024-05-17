package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/rediskey"
	"interview/controllers"
	"interview/helper"
	"interview/models"
	"interview/services"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HomeController struct {
	controllers.Controller
}

type QIDListResp struct {
	List                  []string `json:"list"`
	CurrQID               string   `json:"curr_qid"`
	CurrQIDInAllListIndex int      `json:"curr_qid_in_all_list_index"`
	AllCount              int      `json:"all_count"`
}

func (h *HomeController) QIDList(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		JobTag            string   `json:"job_tag"`
		Provence          string   `json:"provence"`
		QuestionReal      int      `json:"question_real"`
		QuestionCategory  []string `json:"question_category"`
		CurrQuestionID    string   `json:"curr_question_id"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		h.Error(common.CodeInvalidParam, c)
		return
	}
	if param.ExamCategory == "" {
		h.Error(common.CodeInvalidParam, c, "exam_category is empty")
		return
	}

	filter := bson.M{"exam_category": param.ExamCategory}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	filter["question_real"] = param.QuestionReal
	filter["job_tag"] = param.JobTag
	filter["provence"] = param.Provence
	var keypoint models.KeypointStatistic
	err = h.DB().Collection(models.KeypointStatisticTable).Where(filter).Take(&keypoint)
	if err != nil {
		h.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	res := QIDListResp{}
	if len(keypoint.Keypoint) == 0 {
		h.Success(res, c)
		return
	}
	allCate := strings.Join(param.QuestionCategory, "_")
	root := models.KeypointStatisticsResp{
		Child: keypoint.Keypoint,
	}
	allList := make([]string, 0)
	services.FindQIDList(allCate, &root, &allList)
	if len(allList) == 0 {
		h.Success(res, c)
		return
	}
	index := -1
	if param.CurrQuestionID == "" {
		index = 0
	} else {
		for i, v := range allList {
			if v == param.CurrQuestionID {
				index = i
				break
			}
		}
	}
	// 没找到默认返回前 10 个
	if index == -1 {
		end := 10
		if end > len(allList) {
			end = len(allList)
		}
		res.List = allList[:end]
		res.CurrQID = allList[0]
		res.CurrQIDInAllListIndex = 0
		res.AllCount = len(allList)
		h.Success(res, c)
		return
	}
	// 返回前后各 5 个
	start := index - 5
	if start < 0 {
		start = 0
	}
	// 因为切片取值是不包含后一位,所以再加 1
	end := index + 5 + 1
	if end > len(allList) {
		end = len(allList)
	}
	res.List = allList[start:end]
	res.CurrQID = allList[index]
	res.CurrQIDInAllListIndex = index
	res.AllCount = len(allList)
	h.Success(res, c)
}

func (h *HomeController) KeypointStatisticsV2(c *gin.Context) {
	examCategory := c.Query("exam_category")
	examChildCategory := c.Query("exam_child_category")
	jobTag := c.Query("job_tag")
	provence := c.Query("provence")
	questionReal := c.DefaultQuery("question_real", "1") // 1 真题 0 模拟题
	if examCategory == "" {
		h.Error(common.CodeInvalidParam, c, "exam_category is empty")
		return
	}
	var questionRealInt int
	questionRealInt, err := strconv.Atoi(questionReal)
	if err != nil {
		questionRealInt = 1
	}

	var wg sync.WaitGroup
	var keypoint models.KeypointStatistic
	answerMap := make(map[string][]string)
	var keypointErr, answerErr error
	var record map[string]string
	uid := c.GetHeader("X-XTJ-UID")

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1 := time.Now()
		answerMap, answerErr = new(models.UserAnswer).GetUserMap(uid, examCategory, examChildCategory, jobTag, provence, questionRealInt)
		fmt.Println("keypointStatisticsV2 GetUserMap cost:", time.Since(s1).Milliseconds())
	}()
	// 查询keypoint
	wg.Add(1)
	go func() {
		defer wg.Done()
		filter := bson.M{"exam_category": examCategory}
		if examChildCategory != "" {
			filter["exam_child_category"] = examChildCategory
		}
		filter["question_real"] = questionRealInt
		filter["job_tag"] = jobTag
		filter["provence"] = provence
		s2 := time.Now()
		keypointErr = h.DB().Collection(models.KeypointStatisticTable).Where(filter).Take(&keypoint)
		fmt.Println("keypointStatisticsV2 keypoint cost:", time.Since(s2).Milliseconds())
	}()
	// 获取用户流量 record,判断是否有新题
	wg.Add(1)
	go func() {
		defer wg.Done()
		examCate := examCategory
		if examChildCategory != "" {
			examCate += "_" + examChildCategory
		}
		key := fmt.Sprintf(string(rediskey.UserAnswerKeyPointRecord), uid, examCategory, jobTag, provence, questionRealInt)
		record, err = helper.RedisHGetAll(key)
	}()
	wg.Wait()
	if answerErr != nil && !h.MongoNoResult(answerErr) {
		h.Error(common.CodeServerBusy, c, answerErr.Error())
		return
	}
	if keypointErr != nil && !h.MongoNoResult(keypointErr) {
		h.Error(common.CodeServerBusy, c, keypointErr.Error())
		return
	}
	s3 := time.Now()
	if len(keypoint.Keypoint) == 0 {
		h.Success(keypoint.Keypoint, c)
		return
	}
	// 虚拟头部
	root := models.KeypointStatisticsResp{
		Child: keypoint.Keypoint,
	}
	// 把 answer question id  添加到树上
	services.AddAnswerQuestionIDs(&root, answerMap)
	// 子节点 answer count累加到父节点上
	services.DfsAnswerCount(&root)
	// 新增全部 tab
	services.AllTabAnswerCount(&root)
	// last view index 上次浏览的题目
	services.LastViewID(&root, record)
	// 是否有新题目
	services.HasNew(&root, record)
	// 一级分类排序
	services.SortKeypoint(&root)
	fmt.Println("s3:", time.Now().Sub(s3).Milliseconds())
	h.Success(root.Child, c)
}

func (h *HomeController) QuestionAreas(c *gin.Context) {
	var param struct {
		ExamCategory        string `json:"exam_category"`
		ExamChildCategory   string `json:"exam_child_category"`
		IsOnlyShowSource    bool   `json:"is_only_show_source"`
		QuestionReal        int8   `json:"question_real"`         // 是否真题，0为查全部,1真题，2模拟题
		QuestionContentType int8   `json:"question_content_type"` // 试题类别，0普通题，1漫画题, -1所有
		JobTag              string `json:"job_tag"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		h.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"$and": bson.A{
		bson.M{"province": bson.M{"$exists": true, "$ne": ""}},
	}}
	filter["status"] = 5
	if param.QuestionContentType != -1 {
		filter["question_content_type"] = param.QuestionContentType
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
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.IsOnlyShowSource {
		filter["question_source"] = bson.M{"$ne": ""}
	}
	type tmpS struct {
		Provence string `bson:"_id" json:"id"`
		Total    int    `json:"total"`
	}
	var res []tmpS
	err = h.DB().Collection("g_interview_questions").Aggregate(
		bson.A{
			bson.M{"$match": filter},
			bson.M{"$group": bson.M{
				"_id":         "$province",
				"uniqueCount": bson.M{"$addToSet": "$question_source"},
			}},
			bson.M{"$project": bson.M{
				"_id":   1,
				"total": bson.M{"$size": "$uniqueCount"},
			}}}, &res)
	if err != nil {
		h.Error(common.CodeServerBusy, c)
		return
	}
	if len(res) > 1 {
		sort.SliceStable(res, func(i, j int) bool {
			return common.FirstLetterOfPinYin(res[i].Provence) < common.FirstLetterOfPinYin(res[j].Provence)
		})
	}

	h.Success(res, c)
}
