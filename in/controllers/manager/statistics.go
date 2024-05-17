package manager

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mohae/deepcopy"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/controllers"
	"interview/router/request"
	"interview/services"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

/**
* @Author XuPEngHao
* @Date 2023/12/7 12:05
 */

type Statistics struct {
	controllers.Controller
}

func (sf Statistics) QuestionStatistics(c *gin.Context) {
	examCategory := c.Query("exam_category")
	examChildCategory := c.Query("exam_child_category")
	jobTag := c.Query("job_tag")
	province := c.Query("province")
	questionReal, _ := strconv.Atoi(c.Query("question_real")) // 是否真题
	dataType, _ := strconv.Atoi(c.Query("data_type"))

	data, err := services.NewStatisticsSrv().QuestionStatistics(examCategory, examChildCategory, jobTag, province, questionReal, dataType)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

func (sf Statistics) QuestionYearsStatistics(c *gin.Context) {
	examCategory := c.Query("exam_category")
	examChildCategory := c.Query("exam_child_category")
	jobTag := c.Query("job_tag")
	province := c.Query("province")
	questionReal, _ := strconv.Atoi(c.Query("question_real")) // 是否真题

	data, err := services.NewStatisticsSrv().QuestionYearsStatistics(examCategory, examChildCategory, jobTag, province, questionReal)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

// 每日数据
func (sf Statistics) DailyStatistics(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")
	examCategory := c.Query("exam_category")
	examChildCategory := c.Query("exam_child_category")

	filter := bson.M{"created_time": bson.M{"$gt": start, "$lt": end}}
	if examCategory != "" {
		filter["exam_category"] = examCategory
	}
	if examChildCategory != "" {
		filter["exam_child_category"] = examChildCategory
	}

	resp := services.NewStatisticsSrv().DailyUserTestLogStatisticsData(filter)

	sf.Success(resp, c)
}

//	==================================运营需要的每周统计数据=======================================
//
// 查询总练习次数、总练习人数、省份下的练习次数和练习人数
func (sf Statistics) GetStatisticsInfo(c *gin.Context) {
	var err error
	var param struct {
		ExamCategory      string   `json:"exam_category"`       // 问题分类，如事业单位，教招面试，教资面试
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   //题分类
		StartTime         string   `json:"start_time"`
		EndTime           string   `json:"end_time"`
		QuestionIDs       []string `json:"question_ids"`
		QuestionReal      int8     `json:"question_real"` // 是否真题
		JobTag            string   `json:"job_tag"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	filter := bson.M{"log_type": 1}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}

	timeFilter := bson.M{}
	if param.StartTime != "" {
		timeFilter["$gte"] = param.StartTime
	}
	if param.EndTime != "" {
		timeFilter["$lt"] = param.EndTime
	}
	if len(timeFilter) > 0 {
		filter["created_time"] = timeFilter
	}

	if len(param.QuestionIDs) != 0 {
		filter["question_id"] = bson.M{"$in": param.QuestionIDs}
	}
	// =====拼装查询条件=====
	aggregateFilter := bson.A{
		bson.M{"$match": filter},
	}
	secondFilter := bson.M{}
	if param.JobTag != "" {
		secondFilter["question_info.job_tag"] = param.JobTag
	}
	if param.QuestionReal != -1 {
		secondFilter["question_info.question_real"] = param.QuestionReal
	}
	if len(param.QuestionCategory) != 0 {
		secondFilter["question_info.question_category"] = param.QuestionCategory
	}
	if len(secondFilter) > 0 {
		aggregateFilter = append(aggregateFilter,
			bson.M{"$addFields": bson.M{"convertedId": bson.M{"$toObjectId": "$question_id"}}},
			bson.M{"$lookup": bson.M{
				"from":         "g_interview_questions",
				"localField":   "convertedId",
				"foreignField": "_id",
				"as":           "question_info"}})
		aggregateFilter = append(aggregateFilter, bson.M{"$match": secondFilter})
	}

	aggregateFilter = append(aggregateFilter,
		bson.M{"$project": bson.M{
			"user_id":  1,
			"province": 1},
		})
	// =====拼装结束=====

	type tempResp struct {
		UserId   string `bson:"user_id"`
		Province string `bson:"province"`
	}
	var aggregateResp []tempResp
	err = sf.DB().Collection("g_interview_answer_logs").Aggregate(aggregateFilter, &aggregateResp)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	var users []string                                // 用户数
	provinceExerciseInfo := make(map[string]int)      // 省份下的练习次数
	provinceUserCountMap := make(map[string][]string) // 省份下的练习人数
	provinceUserInfo := make(map[string]int)          // 省份下的练习人数
	for _, r := range aggregateResp {
		userId := r.UserId
		province := r.Province
		if !common.InArrCommon(userId, users) {
			users = append(users, userId)
		}

		provinceExerciseInfo[province] += 1

		if !common.InArrCommon(userId, provinceUserCountMap[province]) {
			provinceUserCountMap[province] = append(provinceUserCountMap[province], userId)
			provinceUserInfo[province] += 1
		}
	}

	totalExerciseCount := len(aggregateResp)
	totalUserCount := len(users)
	// 可恶的无序字典！
	// 开始排序！
	type Temp struct {
		Province string `json:"province"`
		Count    int    `json:"count"`
	}
	var pei []Temp
	for k, v := range provinceExerciseInfo {
		pei = append(pei, Temp{k, v})
	}
	sort.Slice(pei, func(i, j int) bool {
		return pei[i].Count > pei[j].Count
	})

	var pui []Temp
	for k, v := range provinceUserInfo {
		pui = append(pui, Temp{k, v})
	}
	sort.Slice(pui, func(i, j int) bool {
		return pui[i].Count > pui[j].Count
	})
	// 排序结束

	sf.Success(map[string]interface{}{
		"total_exercise_count":   totalExerciseCount,
		"total_user_count":       totalUserCount,
		"province_exercise_info": pei,
		"province_user_info":     pui,
	}, c)

}

// 查询单人练习次数和地区信息
func (sf Statistics) GetUserInfoAndExerciseCount(c *gin.Context) {
	var err error
	var param struct {
		ExamCategory      string   `json:"exam_category"`       // 问题分类，如事业单位，教招面试，教资面试
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionCategory  []string `json:"question_category"`   //题分类
		StartTime         string   `json:"start_time"`
		EndTime           string   `json:"end_time"`
		QuestionIDs       []string `json:"question_ids"`
		QuestionReal      int8     `json:"question_real"` // 是否真题
		JobTag            string   `json:"job_tag"`
		Province          string   `json:"province"`
		AboveCount        int      `json:"above_count"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// db.getCollection("g_interview_answer_logs").aggregate([{"$match": {"log_type":1,"created_time": {"$gte":"2024-04-15 00:00:00",}}},
	//{"$group":{"_id":"$user_id", "count": {"$sum":1}}},
	//{"$sort":{"count":-1}},
	//{'$match':{"count":{"$gte":2}}}
	//])

	filter := bson.M{"log_type": 1}
	if param.Province != "" {
		filter["province"] = param.Province
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		filter["exam_child_category"] = param.ExamChildCategory
	}
	if param.ExamCategory != "" {
		filter["exam_category"] = param.ExamCategory
	}

	timeFilter := bson.M{}
	if param.StartTime != "" {
		timeFilter["$gte"] = param.StartTime
	}
	if param.EndTime != "" {
		timeFilter["$lt"] = param.EndTime
	}
	if len(timeFilter) > 0 {
		filter["created_time"] = timeFilter
	}

	if len(param.QuestionIDs) != 0 {
		filter["question_id"] = bson.M{"$in": param.QuestionIDs}
	}
	// =====拼装查询条件=====
	aggregateFilter := bson.A{
		bson.M{"$match": filter},
	}
	// 二次匹配
	secondFilter := bson.M{}
	if param.JobTag != "" {
		secondFilter["question_info.job_tag"] = param.JobTag
	}
	if param.QuestionReal != -1 {
		secondFilter["question_info.question_real"] = param.QuestionReal
	}
	if len(param.QuestionCategory) != 0 {
		secondFilter["question_info.question_category"] = param.QuestionCategory
	}
	if len(secondFilter) > 0 {
		// 这个时候才使用联表查询
		aggregateFilter = append(aggregateFilter,
			bson.M{"$addFields": bson.M{"convertedId": bson.M{"$toObjectId": "$question_id"}}},
			bson.M{"$lookup": bson.M{
				"from":         "g_interview_questions",
				"localField":   "convertedId",
				"foreignField": "_id",
				"as":           "question_info"}})
		aggregateFilter = append(aggregateFilter, bson.M{"$match": secondFilter})
	}

	// 只展示需要的字段
	aggregateFilter = append(aggregateFilter,
		bson.M{"$project": bson.M{
			"user_id": 1,
		},
		})
	// 对user_id进行统计
	aggregateFilter = append(aggregateFilter,
		bson.M{"$group": bson.M{
			"_id":   "$user_id",
			"count": bson.M{"$sum": 1}},
		})
	// 筛选大于指定次数的数据
	aggregateFilter = append(aggregateFilter,
		bson.M{"$match": bson.M{"count": bson.M{"$gte": param.AboveCount}}})
	// 查询符合条件的数量
	aggregateCountFilter := deepcopy.Copy(aggregateFilter)
	var count int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, _ = sf.DB().Collection("g_interview_answer_logs").AggregateCount(aggregateCountFilter.(bson.A))
	}()

	wg.Add(1)
	type tempResp struct {
		UserId string  `bson:"_id" json:"user_id"`
		Count  float64 `bson:"count" json:"count"`
		Area   string  `bson:"-" json:"area"`
	}
	var aggregateResp []tempResp
	go func() {
		defer wg.Done()
		// 对结果进行排序
		aggregateFilter = append(aggregateFilter,
			bson.M{"$sort": bson.M{
				"count": -1,
			},
			})
		// 对结果进行limit
		aggregateFilter = append(aggregateFilter,
			bson.M{"$limit": 50},
		)
		// =====拼装结束=====
		err = sf.DB().Collection("g_interview_answer_logs").Aggregate(aggregateFilter, &aggregateResp)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		// 查询用户地区信息
		var userIds []string
		for _, a := range aggregateResp {
			userIds = append(userIds, a.UserId)
		}
		userAreaMap := sf.getUserAreaFromMysql(userIds)
		for i, a := range aggregateResp {
			aggregateResp[i].Area = userAreaMap[a.UserId]
		}
	}()
	wg.Wait()

	sf.Success(map[string]interface{}{"above_info": count, "exercise_info": aggregateResp}, c)

}

// 从mysql查询用户的地址信息
func (sf Statistics) getUserAreaFromMysql(userIDs []string) map[string]string {
	var userInfoMap = map[string]string{}
	var UserInfo []struct {
		GUID         string `gorm:"column:guid"`
		ProvinceName string `gorm:"column:province_name"`
		CityName     string `gorm:"column:city_name"`
		AdName       string `gorm:"column:ad_name"`
	}
	if len(userIDs) > 0 {
		new(services.ServicesBase).Mysql().Table("xtj_user_info").
			Select("xtj_user_info.guid, xtj_user_info.ad_code, xtj_user_info.ad_code_init, xtj_area.province_name, xtj_area.city_name, xtj_area.ad_name").
			Joins("JOIN xtj_area ON xtj_user_info.ad_code = xtj_area.ad_code").
			Where("xtj_user_info.guid IN ?", userIDs).
			Scan(&UserInfo)
		for _, info := range UserInfo {
			area := info.ProvinceName + "-" + info.CityName + "-" + info.AdName
			userInfoMap[info.GUID] = area
		}
	}
	return userInfoMap
}

func (sf Statistics) GetExerciseRoomCount(c *gin.Context) {
	var err error
	var param struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		ExamType  int8   `json:"exam_type"` // 0 普通考试，1 自主练习
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	filter := bson.M{"exam_type": param.ExamType}
	timeFilter := bson.M{}
	if param.StartTime != "" {
		timeFilter["$gte"] = param.StartTime
	}
	if param.EndTime != "" {
		timeFilter["$lt"] = param.EndTime
	}
	if len(timeFilter) > 0 {
		filter["created_time"] = timeFilter
	}
	count, err := sf.DB().Collection("mock_exam").Where(filter).Count()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]interface{}{"room_count": count}, c)
}

// 数据漏斗
func (sf *Statistics) DataFunneling(c *gin.Context) {
	var err error
	var param request.DataFunnelingRequest
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	// 数据漏斗：
	//面试AI：
	//	打开小程序
	//	刷题
	//	练习场
	//	面试晨读
	response := new(request.DataFunnelingResponse)
	response.Hit = make([]string, 0)
	response.Miss = make([]string, 0)

	guids := strings.Split(param.Guids, "\n")
	if len(guids) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": response, "message": "guids empty"})
		return
	}
	param.GuidsArr = common.RemoveDuplicateElement(guids)
	response.UploadCount = len(param.GuidsArr)
	statisticsService := services.NewStatisticsSrv()
	switch param.Scene {
	case "全部", "打开小程序":
		param.TrackingEvent = "init"
		data, err := statisticsService.GetClickhouseDataFunnelingData(param)
		if err != nil {
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		response.Miss = param.GuidsArr
		response.MissCount = len(response.Miss)
		if len(data) > 0 {
			response.Miss = make([]string, 0)
			response.Hit = data
			for _, datum := range param.GuidsArr {
				if !common.InArrCommon[string](datum, data) {
					response.Miss = append(response.Miss, datum)
				}
			}
			response.HitCount = len(data)
			response.MissCount = len(response.Miss)
		}
	case "刷题", "练习场", "面试晨读":
		data, err := statisticsService.GetDataFunnelingData(param)
		if err != nil {
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		response.Miss = param.GuidsArr
		response.MissCount = len(response.Miss)
		if len(data) > 0 {
			response.Miss = make([]string, 0)
			response.Hit = data
			for _, datum := range param.GuidsArr {
				if !common.InArrCommon[string](datum, data) {
					response.Miss = append(response.Miss, datum)
				}
			}
			response.HitCount = len(data)
			response.MissCount = len(response.Miss)
		}
	default:
		response.Miss = param.GuidsArr
		response.MissCount = len(response.Miss)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": response, "message": "success"})
}
