package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"interview/models"
	"interview/params"
	"interview/router/request"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
)

type MorningReadService struct {
	ServicesBase
}

// 查某天某班的晨读
func (sf *MorningReadService) GetByClassIdAndDate(classId, date string) (models.MorningRead, error) {
	var err error
	read := models.NewMorningRead()
	err = sf.DB().Collection(read.TableName()).Where(bson.M{"class_id": classId, "date": date}).Take(read)
	return *read, err
}

// 删除晨读
func (sf *MorningReadService) DelById(id string) error {
	var err error
	read := models.NewMorningRead()
	_, err = sf.DB().Collection(read.TableName()).Where(bson.M{"_id": sf.ObjectID(id)}).Delete(read)
	return err
}

// 晨读列表
func (sf *MorningReadService) List(filter bson.M, offset, pageSize int64) ([]models.MorningRead, int, error) {
	list := make([]models.MorningRead, 0)
	count, err := sf.DB().Collection(models.NewMorningRead().TableName()).Where(filter).Count()
	if err != nil {
		return list, 0, err
	}
	err = sf.DB().Collection(models.NewMorningRead().TableName()).Where(filter).Skip(offset).Limit(pageSize).Sort("-created_time").Find(&list)

	ids := make([]string, 0, len(list))
	for _, v := range list {
		ids = append(ids, v.Id.Hex())
	}
	logObj := new(models.MorningReadLog)
	type mapData struct {
		MID   string `bson:"_id" json:"mid"`
		Count int64  `bson:"count" json:"count"`
	}
	var logsCount []mapData
	f := bson.A{
		bson.M{"$match": bson.M{"morning_read_id": bson.M{"$in": ids}}},
		bson.M{"$group": bson.M{"_id": "$morning_read_id", "count": bson.M{"$sum": 1}}},
	}
	err = sf.DB().Collection(logObj.TableName()).Aggregate(f, &logsCount)
	if err != nil {
		return list, int(count), err
	}
	countMap := make(map[string]int64)
	for _, v := range logsCount {
		countMap[v.MID] = v.Count
	}
	for i, r := range list {
		if _, ok := countMap[r.Id.Hex()]; !ok {
			continue
		}
		list[i].LogCount = int(countMap[r.Id.Hex()])
		if list[i].InteractiveMode == 0 {
			list[i].InteractiveMode = 1
		}
	}

	return list, int(count), err
}

// 添加/编辑晨读
func (sf *MorningReadService) Save(param request.MorningReadSaveRequest) (string, error) {
	var err error
	read := models.NewMorningRead()
	var isNew = true
	if param.Id != "" {
		err = sf.DB().Collection(read.TableName()).Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(read)
		if err != nil {
			return "", err
		}
		isNew = false
	}
	// 默认封面
	//read.Cover = "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/file-cj4a5sshb88cv6a00s4g"
	//if param.Cover != "" {
	read.Cover = param.Cover
	read.Tags = param.Tags
	//}
	//read.ClassID = param.ClassId
	read.StartDate = param.StartDate
	read.EndDate = param.EndDate
	read.Name = param.Name
	if read.Name == "" {
		read.Name = fmt.Sprintf("%s至%s晨读", read.StartDate, read.EndDate)
		if read.StartDate == read.EndDate {
			read.Name = fmt.Sprintf("%s晨读", read.StartDate)
		}
	}
	read.ManagerId = param.ManagerId
	read.SubName = param.SubName
	read.QuestionAnswer = param.QuestionAnswer
	read.QuestionContent = param.QuestionContent
	read.ReadTimes = param.ReadTimes
	read.Keywords = param.Keywords
	read.Mode = param.Mode
	read.ExamCategory = param.ExamCategory
	//read.ExamChildCategory = param.ExamChildCategory
	read.JobTag = param.JobTag
	read.PreviewBeforeExam = param.PreviewBeforeExam
	read.Sort = param.Sort
	read.InteractiveMode = param.InteractiveMode
	read.ShareImg = param.ShareImg
	read.OpenSilentRead = param.OpenSilentRead
	read.RelationQuestionId = param.RelationQuestionId

	read.Province = param.Province
	if param.State != 0 {
		read.State = param.State
	}
	if read.State == 0 {
		read.State = 1 // 默认上架状态
	}
	if isNew {
		_, err = sf.DB().Collection(read.TableName()).Create(read)
	} else {
		_, err = sf.DB().Collection(read.TableName()).Where(bson.M{"_id": read.Id}).Update(read)
	}

	return read.Id.Hex(), err
}

// 获取今日晨读数据
func (sf *MorningReadService) GetMorningRead(c *gin.Context, offset, pageSize int64) ([]models.MorningRead, error, int64) {
	//var student = models.User{}
	uid := c.GetHeader("X-XTJ-UID")

	morningReadId := c.Query("id")
	exam_category := c.Query("exam_category")
	exam_child_category := c.Query("exam_child_category")
	job_tag := c.Query("job_tag")
	province := c.Query("province")
	tag := c.Query("tag")
	var morningRead = make([]models.MorningRead, 0)
	var err error
	var filter = bson.M{}

	//err = sf.DB().Collection("users").Where(bson.M{"user_id": uid}).Take(&student)
	//if err != nil {
	//	sf.SLogger().Error(err)
	//	return morningRead, err
	//}
	var date = time.Now().Format("2006-01-02")

	if morningReadId != "" {
		filter = bson.M{"_id": sf.ObjectID(morningReadId)}
	} else {
		filter = bson.M{"start_date": bson.M{"$lte": date}, "end_date": bson.M{"$gte": date}, "state": 1}
		filter["$or"] = bson.A{bson.M{"exam_category.0": bson.M{"$exists": false}}}
		examCategoryParam := ""
		if exam_category != "" {
			filter["$or"] = bson.A{bson.M{"exam_category.0": bson.M{"$exists": false}}, bson.M{"exam_category": exam_category}}
			examCategoryParam = exam_category
		}
		if exam_child_category != "" {
			filter["$or"] = bson.A{bson.M{"exam_category.0": bson.M{"$exists": false}}, bson.M{"exam_category": exam_category + "/" + exam_child_category}}
			examCategoryParam = exam_category + "/" + exam_child_category
		}
		if job_tag != "" {
			filter["$or"] = bson.A{bson.M{"exam_category.0": bson.M{"$exists": false}}, bson.M{"exam_category": examCategoryParam, "job_tag": job_tag}}
		}
		if province != "" {
			filter["$or"] = bson.A{bson.M{"exam_category.0": bson.M{"$exists": false}}, bson.M{"exam_category": examCategoryParam, "province": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(province)}}}}
		}
		if tag != "" {
			filter["tags"] = bson.M{"$in": []string{tag}}
		}
	}

	// 看下有没有今天的晨读
	count, _ := sf.DB().Collection(models.NewMorningRead().TableName()).Where(filter).Count()
	err = sf.DB().Collection(models.NewMorningRead().TableName()).Where(filter).Sort("-sort", "-created_time").Skip(offset).Limit(pageSize).Find(&morningRead)
	if err != nil {
		sf.SLogger().Error(err)
	}

	if uid == "" {
		for i := range morningRead {
			morningRead[i] = sf.defaultValue(morningRead[i])
		}
		return morningRead, err, count
	}
	var readIds = make([]string, 0)
	for _, read := range morningRead {
		readIds = append(readIds, read.Id.Hex())
	}
	var mrLog = make([]models.MorningReadLog, 0)
	// 看下晨读记录有没有
	err = sf.DB().Collection(new(models.MorningReadLog).TableName()).Where(bson.M{"morning_read_id": bson.M{"$in": readIds}, "uid": uid}).Find(&mrLog)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		return morningRead, err, 0
	}
	// 没有就生成一个
	if err != nil && sf.MongoNoResult(err) && morningReadId != "" {
		todayMorningLog := new(models.MorningReadLog)
		// 创建一个晨读记录
		todayMorningLog.Uid = uid
		todayMorningLog.Date = time.Now().Format("2006-01-02")
		todayMorningLog.MorningReadId = morningReadId
		_, err = sf.DB().Collection(todayMorningLog.TableName()).Create(todayMorningLog)
	}

	mrLogMap := make(map[string]models.MorningReadLog)
	for _, log := range mrLog {
		mrLogMap[log.MorningReadId] = log
	}

	for i, read := range morningRead {
		if log, ok := mrLogMap[read.Id.Hex()]; ok {
			morningRead[i] = sf.assignLog(morningRead[i], log)
		} else {
			morningRead[i] = sf.defaultValue(morningRead[i])
		}
	}
	return morningRead, err, count
}

// 晨读上报
func (sf *MorningReadService) Report(uid string, param request.MorningReadReportRequest) error {
	var mrLog = models.MorningReadLog{}
	// 看下晨读记录有没有
	err := sf.DB().Collection(mrLog.TableName()).Where(bson.M{"morning_read_id": param.Id, "uid": uid}).Take(&mrLog)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		return err
	}
	var morningRead = models.MorningRead{}
	err2 := sf.DB().Collection(models.NewMorningRead().TableName()).Where(bson.M{"_id": sf.ObjectID(param.Id)}).Take(&morningRead)
	if err2 != nil {
		sf.SLogger().Error(err2)
		return err2
	}
	if sf.MongoNoResult(err) {
		// 创建一个晨读记录
		mrLog.Uid = uid
		mrLog.Date = time.Now().Format("2006-01-02")
		mrLog.MorningReadId = morningRead.Id.Hex()
		mrLog.SourceType = param.SourceType
		_, err = sf.DB().Collection(mrLog.TableName()).Create(&mrLog)
		if err != nil {
			sf.SLogger().Error(err)
			return err
		}
		err = sf.DB().Collection(mrLog.TableName()).Where(bson.M{"morning_read_id": param.Id, "uid": uid}).Take(&mrLog)
		if err != nil {
			sf.SLogger().Error(err)
			return err
		}
	}

	// check 为真是考试模式 FALSE的练习模式
	mrLog.MorningReadIdTags = morningRead.Tags
	mrLog.ReadContent = param.ReadContent
	mrLog.ReadAnswer = param.ReadAnswer
	mrLog.LatestReport = "half_read"
	mrLog.ReadCostTime = param.CostTime
	if param.AllReadReport {
		mrLog.HasReadTimes += 1
		mrLog.LatestReport = "read"
		if morningRead.Mode == 1 && mrLog.HasReadTimes >= morningRead.ReadTimes {
			mrLog.Status = 1
		}
	}
	mrLog.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")

	_, err = sf.DB().Collection(mrLog.TableName()).Where(bson.M{"_id": mrLog.Id}).Update(&mrLog)

	return err
}

// 晨读历史
func (sf *MorningReadService) LogList(filter bson.M, offset, pageSize int64) ([]models.MorningRead, int, error) {
	readList := make([]models.MorningRead, 0)
	//count, err := sf.DB().Collection(new(models.MorningReadLog).TableName()).Where(filter).Count()
	//if err != nil {
	//	sf.SLogger().Error(err)
	//	return readList, 0, err
	//}
	list := make([]models.MorningReadLog, 0)
	err := sf.DB().Collection(new(models.MorningReadLog).TableName()).Where(filter).Sort("-date", "-created_time").Find(&list)
	if err != nil {
		sf.SLogger().Error(err)
	}

	readIds := make([]primitive.ObjectID, 0)
	for _, log := range list {
		readIds = append(readIds, sf.ObjectID(log.MorningReadId))
	}

	err = sf.DB().Collection(models.NewMorningRead().TableName()).Where(bson.M{"_id": bson.M{"$in": readIds}}).Find(&readList)
	if err != nil {
		sf.SLogger().Error(err)
	}
	count := len(readList)
	newReadList := make([]models.MorningRead, 0)

	mrMap := make(map[string]models.MorningRead)
	for _, read := range readList {
		mrMap[read.Id.Hex()] = read
	}
	for _, log := range list {
		mr, ok := mrMap[log.MorningReadId]
		if ok {
			mr = sf.assignLog(mr, log)
			newReadList = append(newReadList, mr)
		}
	}
	pageList := make([]models.MorningRead, 0)
	if len(newReadList) >= int(offset+pageSize) {
		pageList = newReadList[offset:int(offset+pageSize)]
	} else {
		if len(newReadList) > int(offset) {
			pageList = newReadList[offset:]
		}
	}

	return pageList, int(count), err
}

// 晨读历史
func (sf *MorningReadService) ItemLogList(filter bson.M, offset, pageSize int64) ([]models.MorningRead, int, error) {

	list := make([]models.MorningReadLog, 0)
	count, _ := sf.DB().Collection(new(models.MorningReadLog).TableName()).Where(filter).Count()
	err := sf.DB().Collection(new(models.MorningReadLog).TableName()).Where(filter).Sort("-date", "-created_time").Skip(offset).Limit(pageSize).Find(&list)
	if err != nil {
		sf.SLogger().Error(err)
	}
	var readInfo models.MorningRead
	err = sf.DB().Collection(models.NewMorningRead().TableName()).Where(bson.M{"_id": sf.ObjectID(filter["morning_read_id"].(string))}).Take(&readInfo)
	if err != nil {
		sf.SLogger().Error(err)
	}

	pageList := make([]models.MorningRead, 0)

	for _, log := range list {
		mr := readInfo
		mr = sf.assignLog(mr, log)
		pageList = append(pageList, mr)
	}

	uids := make([]string, 0)
	for _, read := range pageList {
		uids = append(uids, read.Uid)
	}

	userGatewayMap := new(User).GetGatewayUsersInfo(uids, "402", 2)
	userMobileMap := new(User).GetMobileInfoFromMysql(uids)

	for i, v := range pageList {
		if user, ok := userGatewayMap[v.Uid]; ok {
			pageList[i].Nickname = user.Nickname
			pageList[i].Avatar = user.Avatar
		}
		if userinfo, ok := userMobileMap[v.Uid]; ok {
			pageList[i].MobileID = userinfo.MobileID
			if pageList[i].MobileID != "" {
				pageList[i].MobileAffiliation = userinfo.Address
			}
		}
	}

	return pageList, int(count), err
}

func (sf *MorningReadService) MorningReadLogList(param params.MorningReadLogParam, offset,
	pageSize int64) ([]params.MorningReadLogResponse, int,
	error) {
	matchFilter := bson.M{}
	keywordFilter := bson.M{}
	aggregateFilter := bson.A{}
	respList := make([]params.MorningReadLogResponse, 0)

	// 晨读时间
	if param.StartTime != "" {
		aggregateFilter = append(aggregateFilter,
			bson.M{"$match": bson.M{"updated_time": bson.M{"$gte": param.StartTime}}},
		)
	}
	if param.EndTime != "" {
		aggregateFilter = append(aggregateFilter,
			bson.M{"$match": bson.M{"updated_time": bson.M{"$lte": param.EndTime}}},
		)
	}
	// 关键词
	if param.Keyword != "" {
		questionContent := param.Keyword
		filterA := bson.A{
			bson.M{"read_content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
			bson.M{"read_answer": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
			bson.M{"morning_read.name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(
				questionContent)}}},
			bson.M{"morning_read.sub_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(
				questionContent)}},
			},
			bson.M{"uid": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
		}
		if len(questionContent) == 24 {
			filterA = append(filterA, bson.M{"_id": sf.ObjectID(questionContent)}, bson.M{"morning_read_id": questionContent})
		}

		keywordFilter = bson.M{"$match": bson.M{"$or": filterA}}
	}
	// 晨读类型
	if param.ExamCategory != "" {
		matchFilter["morning_read.exam_category"] = param.ExamCategory
	}
	// 晨读子类型类型
	if param.ExamChildCategory != "" {
		matchFilter["morning_read.exam_category"] = param.ExamCategory + "/" + param.ExamChildCategory
	}
	// 晨读科目
	if param.JobTag != "" {
		matchFilter["morning_read.job_tag"] = param.JobTag
	}

	if param.Province != "" {
		matchFilter["morning_read.province"] = param.Province
	}
	// 作答学员id
	if param.Uid != "" {
		aggregateFilter = append(aggregateFilter,
			bson.M{"$match": bson.M{"uid": param.Uid}},
		)
	}

	aggregateFilterBson := bson.A{
		bson.M{"$addFields": bson.M{"convertedId": bson.M{"$toObjectId": "$morning_read_id"}}},
		bson.M{"$lookup": bson.M{
			"from":         "morning_read",
			"localField":   "convertedId",
			"foreignField": "_id",
			"as":           "morning_read"},
		},
	}
	if len(aggregateFilter) != 0 {
		aggregateFilter = append(aggregateFilter, aggregateFilterBson...)
	} else {
		aggregateFilter = aggregateFilterBson
	}
	if len(keywordFilter) != 0 {
		aggregateFilter = append(aggregateFilter, keywordFilter)
	}
	if len(matchFilter) != 0 {
		aggregateFilter = append(aggregateFilter, bson.M{"$match": matchFilter})
	}
	list := make([]params.MorningReadLogResponseParam, 0)
	count, _ := sf.DB().Collection(new(models.MorningReadLog).TableName()).AggregateCount(aggregateFilter)
	// 排序 分页相关
	aggregateFilter = append(aggregateFilter,
		bson.M{"$sort": bson.M{"updated_time": -1}},
		bson.M{"$skip": offset},
		bson.M{"$limit": pageSize},
		bson.M{"$project": bson.M{
			"_id":             1,
			"updated_time":    1,
			"check_answer":    1,
			"check_score":     1,
			"check_cost_time": 1,
			"morning_read":    1,
			"read_cost_time":  1,
			"has_read_times":  1,
			"read_content":    1,
			"read_answer":     1,
			"latest_report":   1,
			"uid":             1,
			"checked":         1,
		},
		})
	timeNow := time.Now()
	sf.SLogger().Info("read log mongo start............")
	err := sf.DB().Collection(new(models.MorningReadLog).TableName()).Aggregate(aggregateFilter, &list)
	if err != nil {
		sf.SLogger().Error(err)
	}
	costTime := time.Now().Sub(timeNow).Seconds()
	msg := fmt.Sprintf("read log mongo end............cost_time=%f", costTime)
	sf.SLogger().Info(msg)
	userIdList := make([]string, 0)
	for _, responseParam := range list {
		serviceParam := params.MorningReadServiceParamList{}
		userIdList = append(userIdList, responseParam.Uid)
		if len(responseParam.MorningReadParam) > 0 {
			serviceParam = responseParam.MorningReadParam[0]
		}
		if len(responseParam.ReadAnswer) == 0 && len(responseParam.CheckAnswer) == 0 {
			responseParam.ReadTime = "-"
		}
		if len(responseParam.ReadContent) == 0 {
			responseParam.ReadContent = serviceParam.QuestionContent
		}
		respList = append(respList, params.MorningReadLogResponse{
			MorningReadLogResponseParam: responseParam,
			MorningReadServiceParamList: serviceParam,
		})
	}
	// 请求用户信息
	sf.SLogger().Info("read log user start............")
	timeUserNow := time.Now()
	userGatewayMap := new(User).GetGatewayUsersInfo(userIdList, "402", 2)
	for i, v := range respList {
		if user, ok := userGatewayMap[v.Uid]; ok {
			respList[i].Nickname = user.Nickname
		}
	}
	costTime = time.Now().Sub(timeUserNow).Seconds()
	msg = fmt.Sprintf("read log user end............cost_time=%f", costTime)
	sf.SLogger().Info(msg)
	return respList, int(count), err
}

// 获取报告
func (sf *MorningReadService) GetMorningReadReport(uid, morningReadId string) (models.MorningRead, error) {
	morningRead := models.MorningRead{}
	err := sf.DB().Collection(morningRead.TableName()).Where(bson.M{"_id": sf.ObjectID(morningReadId)}).Take(&morningRead)
	if err != nil {
		return morningRead, err
	}

	morningReadLog := new(models.MorningReadLog)
	err = sf.DB().Collection(morningReadLog.TableName()).Where(bson.M{"uid": uid, "morning_read_id": morningReadId}).Take(morningReadLog)
	if err != nil {
		return morningRead, err
	}

	morningRead = sf.assignLog(morningRead, *morningReadLog)

	return morningRead, nil
}

func (sf *MorningReadService) assignLog(morningRead models.MorningRead, log models.MorningReadLog) models.MorningRead {
	var emptySlice = make([]string, 0)

	morningRead.HasReadTimes = log.HasReadTimes
	morningRead.Checked = log.Checked
	morningRead.ReadCostTime = log.ReadCostTime
	morningRead.LogCreatedTime = log.CreatedTime
	morningRead.Uid = log.Uid

	readContent := log.ReadContent
	if len(readContent) == 0 {
		readContent = emptySlice
	}
	morningRead.ReadContent = readContent

	readAnswer := log.ReadAnswer
	if len(readAnswer) == 0 {
		readAnswer = emptySlice
	}
	morningRead.ReadAnswer = readAnswer

	checkAnswerResult := log.CheckAnswerResult
	if len(checkAnswerResult) == 0 {
		checkAnswerResult = make([]models.CheckAnswerResult, 0)
	}
	morningRead.CheckAnswerResult = checkAnswerResult

	checkContent := log.CheckContent
	if len(checkContent) == 0 {
		checkContent = emptySlice
	}
	morningRead.CheckContent = checkContent

	checkAnswer := log.CheckAnswer
	if len(checkAnswer) == 0 {
		checkAnswer = emptySlice
	}
	morningRead.CheckAnswer = checkAnswer

	morningRead.CheckScore = log.CheckScore
	morningRead.CheckCostTime = log.CheckCostTime
	morningRead.CheckKeywordsNum = log.CheckKeywordsNum
	morningRead.CheckMatchKeywordsNum = log.CheckMatchKeywordsNum
	morningRead.LatestReport = log.LatestReport
	if morningRead.InteractiveMode == 0 {
		morningRead.InteractiveMode = 1
	}

	return morningRead
}

func (sf *MorningReadService) defaultValue(morningRead models.MorningRead) models.MorningRead {
	var emptySlice = make([]string, 0)
	if len(morningRead.ReadAnswer) == 0 {
		morningRead.ReadAnswer = emptySlice
	}
	if len(morningRead.ReadContent) == 0 {
		morningRead.ReadContent = emptySlice
	}
	if len(morningRead.CheckAnswer) == 0 {
		morningRead.CheckAnswer = emptySlice
	}
	if len(morningRead.CheckContent) == 0 {
		morningRead.CheckContent = emptySlice
	}
	if len(morningRead.CheckAnswerResult) == 0 {
		morningRead.CheckAnswerResult = make([]models.CheckAnswerResult, 0)
	}
	if morningRead.InteractiveMode == 0 {
		morningRead.InteractiveMode = 1
	}

	return morningRead
}
