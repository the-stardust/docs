package services

import (
	"encoding/json"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/common/global"
	"interview/models"
	"interview/models/managerresp"
	"interview/router/request"
	"sort"
	"strings"
	"sync"
	"time"
)

/**
* @Author XuPEngHao
* @Date 2023/12/7 12:10
 */
type StatisticsSrv struct {
	ServicesBase
}

func NewStatisticsSrv() *StatisticsSrv {
	return new(StatisticsSrv)
}

func (sf *StatisticsSrv) QuestionYearsStatistics(examCategory, examChildCategory, jobTag, province string, questionReal int) ([]managerresp.QuestionYearsStatisticsResp, error) {
	baseFilterList := sf.getQuesStatBaseFilter(examCategory, examChildCategory, jobTag, province, questionReal)

	// 获取该身下面的分类
	quesCategoryList, _ := NewCategorySrv().GetQuestionCategoryItem(examCategory, examChildCategory)
	quesCategoryTitleList := make([]string, 0)
	for _, val := range quesCategoryList {
		if val.Title == "" {
			continue
		}
		quesCategoryTitleList = append(quesCategoryTitleList, val.Title)
	}
	quesCategoryTitleList = append(quesCategoryTitleList, "其他(无试题分类)")
	// 并发获取该分类下面的年份数据
	quesCategoryStaticList := make([]managerresp.QuestionYearsStatisticsResp, len(quesCategoryTitleList))
	var wg sync.WaitGroup

	antsPool, _ := ants.NewPool(len(quesCategoryTitleList))
	defer antsPool.Release()
	for i, val := range quesCategoryTitleList {
		wg.Add(1)
		quesCategoryStaticList[i].QuestionCategory = val
		antsPool.Submit(func(wg *sync.WaitGroup, baseFilterList bson.A, index int) func() {
			return func() {
				sf.getQuesCatYestList(wg, baseFilterList, quesCategoryStaticList, index)
			}
		}(&wg, baseFilterList, i))
	}
	wg.Wait()

	return quesCategoryStaticList, nil
}

func (sf *StatisticsSrv) getQuesStatBaseFilter(examCategory, examChildCategory, jobTag, province string, questionReal int) bson.A {
	baseFilterList := bson.A{
		bson.M{"exam_category": examCategory},
	}
	if examChildCategory != "" {
		baseFilterList = append(baseFilterList, bson.M{"exam_child_category": examChildCategory})
	}
	if jobTag != "" {
		baseFilterList = append(baseFilterList, bson.M{"job_tag": jobTag})
	}
	if province != "" {
		if province == "其他（暂无省份）" {
			province = ""
		}
		baseFilterList = append(baseFilterList, bson.M{"province": province})
	}
	if questionReal > -1 {
		baseFilterList = append(baseFilterList, bson.M{"question_real": questionReal})
	}
	return baseFilterList
}

func (sf *StatisticsSrv) getQuesCatYestList(wg *sync.WaitGroup, baseFilterList bson.A, quesCategoryStaticList []managerresp.QuestionYearsStatisticsResp, index int) {
	defer wg.Done()
	type TempStat struct {
		Id    int64 `bson:"_id" json:"id"`
		Count int64 `json:"count"`
	}
	statList := make([]TempStat, 0)

	filterList := bson.A{}
	filterList = append(filterList, baseFilterList...)
	if quesCategoryStaticList[index].QuestionCategory == "其他(无试题分类)" {
		filterList = append(filterList,
			bson.M{"$or": bson.A{
				bson.M{"question_category": bson.M{"$exists": false}},
				bson.M{"question_category": bson.M{"$eq": nil}},
				bson.M{"question_category": bson.M{"$eq": make([]string, 0)}},
			}},
		)
	} else {
		filterList = append(filterList, bson.M{"question_category": quesCategoryStaticList[index].QuestionCategory})
	}

	m := bson.A{
		bson.M{"$match": bson.M{"$and": filterList}},
		bson.M{"$group": bson.M{"_id": "$year", "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"_id": -1}},
	}
	err := sf.DB().Collection((&models.GQuestion{}).TableName()).Aggregate(m, &statList)
	if err != nil {
		sf.SLogger().Errorf("getQuesCatYestList.statList.err=%s", err)
	}

	// 上架
	filterList = append(filterList, bson.M{"status": 5})
	statOnlineList := make([]TempStat, 0)
	err = sf.DB().Collection((&models.GQuestion{}).TableName()).Aggregate(bson.A{
		bson.M{"$match": bson.M{"$and": filterList}},
		bson.M{"$group": bson.M{"_id": "$year", "count": bson.M{"$sum": 1}}},
		bson.M{"$sort": bson.M{"_id": -1}},
	}, &statOnlineList)
	if err != nil {
		sf.SLogger().Errorf("getQuesCatYestList.statOnlineList.err=%s", err)
	}
	statOnlineMap := make(map[int64]int64)
	for _, val := range statOnlineList {
		statOnlineMap[val.Id] = val.Count
	}

	yearDetailList := make([]managerresp.QuestionYearsDetail, len(statList))
	for i, val := range statList {
		yearDetailList[i] = managerresp.QuestionYearsDetail{
			Years:     val.Id,
			Total:     val.Count,
			OnlineCnt: statOnlineMap[val.Id],
		}
	}
	quesCategoryStaticList[index].List = yearDetailList
}

func (sf *StatisticsSrv) QuestionStatistics(examCategory, examChildCategory, jobTag, province string, questionReal, dataType int) (managerresp.QuestionStatisticsResp, error) {
	baseFilterList := sf.getQuesStatBaseFilter(examCategory, examChildCategory, jobTag, province, questionReal)
	result := managerresp.QuestionStatisticsResp{}
	switch dataType {
	case 1:
		result.Summary = sf.getQuesSummary(baseFilterList)
	case 2:
		result.List = sf.GetQuesDetailList(baseFilterList, examCategory, examChildCategory, province)
	}
	return result, nil
}

func (sf *StatisticsSrv) GetQuesDetailList(baseFilterList bson.A, examCategory, examChildCategory, province string) []managerresp.QuestionStatisticsDetail {
	groupId := bson.M{}
	if province == "" {
		groupId["province"] = "$province"
	}
	groupId["question_category"] = bson.M{"$first": "$question_category"}
	newFilterList := bson.A{
		bson.M{"$match": bson.M{"$and": baseFilterList}},
		bson.M{"$group": bson.M{"_id": groupId, "count": bson.M{"$sum": 1}}},
	}

	type ProvinceWithCat struct {
		Province         string `bson:"province" json:"province"`
		QuestionCategory string `bson:"question_category" json:"question_category"`
	}

	// 总数量
	type TempDetailStat struct {
		Id    ProvinceWithCat `bson:"_id" json:"id"`
		Count int64           `bson:"count" json:"count"`
	}
	statList := make([]TempDetailStat, 0)
	err := sf.DB().Collection((&models.GQuestion{}).TableName()).Aggregate(newFilterList, &statList)
	if err != nil {
		sf.SLogger().Errorf("err=%s", err)
		return make([]managerresp.QuestionStatisticsDetail, 0)
	}

	var provinceList []string

	type Temp struct {
		Count       int64
		CategoryMap map[string]int64
	}

	totalMap := make(map[string]Temp)
	for _, val := range statList {
		provinceVal := val.Id.Province
		questionCategoryVal := val.Id.QuestionCategory

		tInfo, ok := totalMap[provinceVal]
		if !ok {
			provinceList = append(provinceList, provinceVal) // 添加省
			tInfo = Temp{Count: 0, CategoryMap: make(map[string]int64)}
		}
		tInfo.Count += val.Count // 总total
		tInfo.CategoryMap[questionCategoryVal] = val.Count
		totalMap[provinceVal] = tInfo
	}

	// 上架数量
	newFilterList = bson.A{
		bson.M{"$match": bson.M{"$and": append(baseFilterList, bson.M{"status": 5})}},
		bson.M{"$group": bson.M{"_id": groupId, "count": bson.M{"$sum": 1}}},
	}
	statList2 := make([]TempDetailStat, 0)
	err = sf.DB().Collection((&models.GQuestion{}).TableName()).Aggregate(newFilterList, &statList2)
	if err != nil {
		sf.SLogger().Errorf("err=%s", err)
		return make([]managerresp.QuestionStatisticsDetail, 0)
	}

	totalOnlineMap := make(map[string]Temp)
	for _, val := range statList2 {
		provinceVal := val.Id.Province
		questionCategoryVal := val.Id.QuestionCategory
		tInfo, ok := totalOnlineMap[provinceVal]
		if !ok {
			tInfo = Temp{Count: 0, CategoryMap: make(map[string]int64)}
		}
		tInfo.Count += val.Count // 总total
		tInfo.CategoryMap[questionCategoryVal] = val.Count
		totalOnlineMap[provinceVal] = tInfo
	}

	quesCategoryList, _ := NewCategorySrv().GetQuestionCategoryItem(examCategory, examChildCategory)
	quesCategoryTitleList := make([]string, 0)
	for _, val := range quesCategoryList {
		if val.Title == "" {
			continue
		}
		quesCategoryTitleList = append(quesCategoryTitleList, val.Title)
	}
	quesCategoryTitleList = append(quesCategoryTitleList, "")

	// 排序provinceList
	detailList := make([]managerresp.QuestionStatisticsDetail, 0)
	for _, p := range provinceList {
		detail := managerresp.QuestionStatisticsDetail{
			Title:        p,
			Total:        0,
			OnlineCnt:    0,
			CategoryList: nil,
		}
		if detail.Title == "" {
			detail.Title = "其他（暂无省份）"
		}
		if province != "" {
			detail.Title = province
		}
		totalInfo, ok2 := totalMap[p]
		if !ok2 {
			totalInfo = Temp{Count: 0, CategoryMap: make(map[string]int64)}
		}

		totalOnlineInfo, ok3 := totalOnlineMap[p]
		if !ok3 {
			totalOnlineInfo = Temp{Count: 0, CategoryMap: make(map[string]int64)}
		}

		detail.Total = totalInfo.Count
		detail.OnlineCnt = totalOnlineInfo.Count

		// 分类计算
		for _, categoryName := range quesCategoryTitleList {
			title := categoryName
			if title == "" {
				title = "其他(无试题分类)"
			}
			detail.CategoryList = append(detail.CategoryList, managerresp.QuestionStatisticsCategory{
				Title:     title,
				Total:     totalInfo.CategoryMap[categoryName],
				OnlineCnt: totalOnlineInfo.CategoryMap[categoryName],
			})
		}

		detailList = append(detailList, detail)
	}

	if len(detailList) > 1 {
		sort.SliceStable(detailList, func(i, j int) bool {
			return common.FirstLetterOfPinYin(detailList[i].Title) < common.FirstLetterOfPinYin(detailList[j].Title)
		})
	}
	return detailList
}

func (sf *StatisticsSrv) getQuesSummary(baseFilterList bson.A) managerresp.Summary {
	result := managerresp.Summary{}
	// 总数量
	result.Total, _ = sf.DB().Collection((&models.GQuestion{}).TableName()).Where(bson.M{"$and": baseFilterList}).Count()

	// 上架数量
	baseFilterList = append(baseFilterList, bson.M{"status": 5})
	result.OnlineCnt, _ = sf.DB().Collection((&models.GQuestion{}).TableName()).Where(bson.M{"$and": baseFilterList}).Count()

	result.OfflineCnt = result.Total - result.OnlineCnt
	return result
}

func (sf *StatisticsSrv) DailyUserTestLogStatisticsData(filter bson.M) any {
	galds := make([]models.GAnswerLogDailyStatistics, 0)
	sf.GAnswerLogDailyStatisticsModel().Where(filter).Find(&galds)
	return galds
}

func (sf *StatisticsSrv) DailyUserTestLogStatistics2(filter bson.M) {
	logs := make([]models.GAnswerLog, 0)
	sf.GAnswerLogModel().Where(filter).Find(&logs)

	selfPractice := make(map[string]*models.GAnswerLogDailyStatistics)
	userIds := make([]string, 0)
	defaultS := struct{}{}
	reviewCount := make(map[string]*models.GAnswerLogDailyStatistics)
	for _, log := range logs {
		userIds = append(userIds, log.UserId)
	}
	logedUserId := make(map[string]map[string]struct{}, 0)
	// 看下是否是学生
	usermap, _ := NewUser().GetUserMap(userIds)

	questionIds := make([]string, 0)
	for _, log := range logs {
		questionIds = append(questionIds, log.QuestionId)
	}
	if len(questionIds) == 0 {
		return
	}
	Gquestions := make([]models.GQuestion, 0)
	GquestionMap := make(map[string]models.GQuestion, 0)
	sf.GQuestionModel().Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(questionIds)}}).Find(&Gquestions)
	for _, gquestion := range Gquestions {
		GquestionMap[gquestion.Id.Hex()] = gquestion
	}
	for _, log := range logs {
		if _, ok := GquestionMap[log.QuestionId]; !ok {
			continue
		}
		log.ExamCategory = GquestionMap[log.QuestionId].ExamCategory
		log.ExamChildCategory = GquestionMap[log.QuestionId].ExamChildCategory
		log.QuestionCategory = GquestionMap[log.QuestionId].QuestionCategory
		log.JobTag = GquestionMap[log.QuestionId].JobTag
		_, userok := usermap[log.UserId]
		statisticType := "category"
		sf.formatDailyUserTestLogStatistics2(reviewCount, log, userok, statisticType, logedUserId)
		if _, ok := logedUserId[log.ExamCategory+log.ExamChildCategory+statisticType]; !ok {
			logedUserId[log.ExamCategory+log.ExamChildCategory+statisticType] = make(map[string]struct{})
		}
		logedUserId[log.ExamCategory+log.ExamChildCategory+statisticType][log.UserId] = defaultS
		if log.JobTag != "" {
			statisticType = "job_tag"
			sf.formatDailyUserTestLogStatistics2(selfPractice, log, userok, statisticType, logedUserId)
			logedUserId[log.ExamCategory+log.ExamChildCategory+statisticType][log.UserId] = defaultS
		}
	}
	_ = sf.saveDailyUserTestLogStatistics(reviewCount)
	_ = sf.saveDailyUserTestLogStatistics(selfPractice)
}

// formatDailyUserTestLogStatistics2
func (sf *StatisticsSrv) formatDailyUserTestLogStatistics2(reviewCount map[string]*models.GAnswerLogDailyStatistics, log models.GAnswerLog, isClassUser bool, statisticType string, logedUserId map[string]map[string]struct{}) {
	examCategory := log.ExamCategory + log.ExamChildCategory + statisticType
	if _, ok := reviewCount[examCategory]; !ok {
		reviewCount[examCategory] = new(models.GAnswerLogDailyStatistics)
		reviewCount[examCategory].Date = log.CreatedTime[:10]
		reviewCount[examCategory].LogType = statisticType
		reviewCount[examCategory].ExamCategory = log.ExamCategory
		reviewCount[examCategory].ExamChildCategory = log.ExamChildCategory
		reviewCount[examCategory].SubjectCategories = make(map[string]*models.GAnswerLogDailyStatistics)
	}
	if _, ok := logedUserId[examCategory]; !ok {
		logedUserId[examCategory] = make(map[string]struct{})
	}
	reviewCount[examCategory].LogCount++
	// 地面班还是自然流量
	if isClassUser {
		reviewCount[examCategory].ClassLogCount++
	} else {
		reviewCount[examCategory].OnlineLogCount++
	}
	// 人数
	if _, ok := logedUserId[examCategory][log.UserId]; !ok {
		reviewCount[examCategory].UserCount++
		if isClassUser {
			reviewCount[examCategory].ClassUserCount++
		} else {
			reviewCount[examCategory].OnlineUserCount++
		}
	}

	// 人数
	if _, ok := logedUserId[examCategory][log.UserId]; !ok {
		reviewCount[examCategory].CompleteUserCount++

	}

	subjectCategory := ""
	if "job_tag" == statisticType {
		subjectCategory = log.JobTag
	} else if len(log.QuestionCategory) > 0 {
		subjectCategory = log.QuestionCategory[0]
	}

	if _, ok := reviewCount[examCategory].SubjectCategories[subjectCategory]; !ok {
		reviewCount[examCategory].SubjectCategories[subjectCategory] = new(models.GAnswerLogDailyStatistics)
		reviewCount[examCategory].SubjectCategories[subjectCategory].ExamCategory = log.ExamCategory
		reviewCount[examCategory].SubjectCategories[subjectCategory].ExamChildCategory = log.ExamChildCategory
		reviewCount[examCategory].SubjectCategories[subjectCategory].QuestionCategory = subjectCategory
		reviewCount[examCategory].SubjectCategories[subjectCategory].SubjectCategories = make(map[string]*models.GAnswerLogDailyStatistics)
	}
	reviewCount[examCategory].SubjectCategories[subjectCategory].LogCount++
	// 地面班还是自然流量
	if isClassUser {
		reviewCount[examCategory].SubjectCategories[subjectCategory].ClassLogCount++
	} else {
		reviewCount[examCategory].SubjectCategories[subjectCategory].OnlineLogCount++
	}

	if _, ok := logedUserId[examCategory][log.UserId]; !ok {
		reviewCount[examCategory].SubjectCategories[subjectCategory].CompleteUserCount++
	}

	// 人数
	if _, ok := logedUserId[examCategory][log.UserId]; !ok {
		reviewCount[examCategory].SubjectCategories[subjectCategory].UserCount++
		if isClassUser {
			reviewCount[examCategory].SubjectCategories[subjectCategory].ClassUserCount++
		} else {
			reviewCount[examCategory].SubjectCategories[subjectCategory].OnlineUserCount++
		}
	}
}

func (sf *StatisticsSrv) saveDailyUserTestLogStatistics(statistics map[string]*models.GAnswerLogDailyStatistics) error {
	var err error
	for _, logStatistics := range statistics {
		filter := bson.M{
			"date":                logStatistics.Date,
			"log_type":            logStatistics.LogType,
			"exam_category":       logStatistics.ExamCategory,
			"exam_child_category": logStatistics.ExamChildCategory,
		}
		ls := new(models.GAnswerLogDailyStatistics)
		err = sf.GAnswerLogDailyStatisticsModel().Where(filter).Take(ls)
		if sf.MongoNoResult(err) {
			_, err = sf.GAnswerLogDailyStatisticsModel().Create(logStatistics)
		} else {
			_, err = sf.GAnswerLogDailyStatisticsModel().Where(bson.M{"_id": ls.Id}).Update(logStatistics)
		}
	}
	return err
}

func (sf *StatisticsSrv) GetClickhouseDataFunnelingData(param request.DataFunnelingRequest) ([]string, error) {
	condition := ""
	if param.StartDate != "" {
		startT, err := time.ParseInLocation("2006-01-02 15:04:05", param.StartDate+" 00:00:00", time.Local)
		if err != nil {
			sf.SLogger().Error(err)
			return nil, err
		}
		endT, err := time.ParseInLocation("2006-01-02 15:04:05", param.EndDate+" 23:59:59", time.Local)
		if err != nil {
			sf.SLogger().Error(err)
			return nil, err
		}

		condition = fmt.Sprintf("event='%s' and created BETWEEN %d and %d and distinct_id in ('%s')", param.TrackingEvent, startT.Unix(), endT.Unix(), strings.Join(param.GuidsArr, "','"))
	} else {
		condition = fmt.Sprintf("event='%s' and distinct_id in ('%s')", param.TrackingEvent, strings.Join(param.GuidsArr, "','"))
	}
	rp := request.ClickhouseTrackingRequest{
		Sign:           "xxx",
		Timestamp:      time.Now().Unix(),
		TableId:        -1,
		Table:          "interviewai_qall",
		Column:         "",
		RespType:       7,
		Condition:      condition,
		ConditionExtra: "",
		OrderBy:        "",
		PageSize:       10,
		Page:           1,
	}

	res, err := common.HttpPostJson(global.CONFIG.ServiceUrls.ClickhouseStatisticUrl+"/question-bank/v1/click/summary", rp)
	if err != nil {
		sf.SLogger().Error("GetClickhouseData err:", err, " resp:", string(res))
		return nil, err
	}
	sf.SLogger().Info("condition: ", condition, "GetClickhouseData: ", string(res))
	type TempData string
	type TempRes struct {
		Msg  string `json:"message"`
		Code int    `json:"code"`
		Data struct {
			List []string `json:"list"`
		} `json:"data"`
	}
	r := TempRes{}
	err = json.Unmarshal(res, &r)
	if err != nil || r.Code != 0 {
		sf.SLogger().Error(err)
		return nil, err
	}

	return r.Data.List, nil
}

func (sf *StatisticsSrv) GetDataFunnelingData(param request.DataFunnelingRequest) ([]string, error) {
	filter := bson.M{"user_id": bson.M{"$in": param.GuidsArr}}
	var uids = make([]string, 0)
	if param.Scene == "刷题" {
		if param.StartDate != "" {
			filter["created_time"] = bson.M{"$gte": param.StartDate + " 00:00:00", "$lte": param.EndDate + " 23:59:59"}
		}
		var logs = make([]models.GAnswerLog, 0)
		sf.GAnswerLogModel().Where(filter).Find(&logs)
		for _, log := range logs {
			uids = append(uids, log.UserId)
		}
	} else if param.Scene == "练习场" {
		if param.StartDate != "" {
			filter["created_time"] = bson.M{"$gte": param.StartDate + " 00:00:00", "$lte": param.EndDate + " 23:59:59"}
		}
		var logs = make([]models.MockExamLog, 0)
		sf.MockExamLogModel().Where(filter).Find(&logs)
		for _, log := range logs {
			uids = append(uids, log.UserId)
		}
	} else if param.Scene == "面试晨读" {
		filter := bson.M{"uid": bson.M{"$in": param.GuidsArr}}
		if param.StartDate != "" {
			filter["created_time"] = bson.M{"$gte": param.StartDate + " 00:00:00", "$lte": param.EndDate + " 23:59:59"}
		}
		var logs = make([]models.MorningReadLog, 0)
		sf.GAnswerLogModel().Where(filter).Find(&logs)
		for _, log := range logs {
			uids = append(uids, log.Uid)
		}
	}
	return common.RemoveDuplicateElement(uids), nil
}
