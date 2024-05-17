package manager

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/ffmt.v1"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"strings"
	"time"
)

type VideoKeypoint struct {
	controllers.Controller
}

var IdentifyInterviewStr = "this is interview, its' type is 1"
var IdentifyQuestionStr = "this is question-with-qk"

func (sf *VideoKeypoint) VideoKeypointCallback(c *gin.Context) {
	var param struct {
		Time        int    `json:"time"`
		TimeEnd     int    `json:"timeEnd"`
		Desc        string `json:"desc"`
		Url         string `json:"url"`
		JiangjieKey string `json:"jiangjieKey"`
		FileId      string `json:"file_id"`
		DataId      string `json:"dataID"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	isSyncVKRelevance := true
	tableName := "video_keypoints"
	if strings.Contains(param.Desc, IdentifyQuestionStr) {
		tableName = "video_questions"
		isSyncVKRelevance = false
	}
	// 保存信息
	updateContent := bson.M{
		"updated_time":      time.Now().Format("2006-01-02 15:04:05"),
		"start_time":        param.Time,
		"end_time":          param.TimeEnd,
		"jiangjie_key":      param.JiangjieKey,
		"desc":              param.Desc,
		"video_url":         param.Url,
		"file_id":           param.FileId,
		"question_category": strings.Split(param.JiangjieKey, "/"),
	}
	_, err = sf.DB().Collection(tableName).Where(bson.M{"_id": sf.ObjectID(param.DataId)}).Update(updateContent)
	if err != nil {
		sf.SLogger().Error(err)
		sf.SLogger().Error("param.data_id:", param.DataId)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 同步更新关联表
	if isSyncVKRelevance {
		updateContent = bson.M{
			"updated_time": time.Now().Format("2006-01-02 15:04:05"),
			"start_time":   param.Time,
			"end_time":     param.TimeEnd,
			"jiangjie_key": param.JiangjieKey,
			"desc":         param.Desc,
			"video_url":    param.Url,
			"file_id":      param.FileId,
		}
		_, err = sf.DB().Collection("video_keypoints_with_relevancy").Where(bson.M{"vk_id": param.DataId}).Update(updateContent)
		if err != nil {
			sf.SLogger().Error(err)
			sf.SLogger().Error("param.data_id:", param.DataId)
		}
	}
	sf.Success(nil, c)
}

func (sf *VideoKeypoint) SaveVideoKeypoints(c *gin.Context) {
	type paramStruct struct {
		Id                string   `json:"id"`
		IsDelete          bool     `json:"is_delete"`           // 是否为删除操作
		ExamCategory      string   `json:"exam_category"`       //考试分类
		ExamChildCategory string   `json:"exam_child_category"` //考试子分类
		QuestionKeypoints []string `json:"question_keypoints"`  //题知识点分类
		StartTime         int      `json:"start_time"`
		EndTime           int      `json:"end_time"`
		Desc              string   `json:"desc"`
		JiangjieKey       string   `json:"jiangjie_key"`
		Type              int8     `json:"type"`        // 0是笔试AI，1是面试AI
		QuestionID        string   `json:"question_id"` // 试题id，仅试题关联视频时才需要传
	}

	var param struct {
		MediaId             string        `json:"media_id" bson:"media_id"`   // 节详情里的Vid
		LessonId            int           `json:"lesson_id" bson:"lesson_id"` // lesson_id
		IsRelevanceQuestion bool          `json:"is_relevance_question"`      // 是否是关联试题的打点
		Infos               []paramStruct `json:"infos" bson:"infos"`         // 打点信息slice
	}
	ffmt.Puts(param)
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	tableName := "video_keypoints"
	if param.IsRelevanceQuestion {
		tableName = "video_questions"
	}
	var bulkSlice []mongo.WriteModel
	var vkRelevanceBulkSlice []mongo.WriteModel
	for _, info := range param.Infos {
		if info.IsDelete {
			// 删除打点信息
			bulkSlice = append(bulkSlice, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": sf.ObjectID(info.Id)}))
			vkRelevanceBulkSlice = append(vkRelevanceBulkSlice, mongo.NewDeleteOneModel().SetFilter(bson.M{"exam_category": info.ExamCategory, "exam_child_category": info.ExamChildCategory, "question_category": info.QuestionKeypoints, "vk_id": info.Id}))
		} else {
			// 更改打点信息
			if info.Id != "" {
				info.Desc = strings.ReplaceAll(info.Desc, IdentifyQuestionStr, "")
				updateContent := bson.M{
					"updated_time":        time.Now().Format("2006-01-02 15:04:05"),
					"start_time":          info.StartTime,
					"end_time":            info.EndTime,
					"question_category":   info.QuestionKeypoints,
					"jiangjie_key":        info.JiangjieKey,
					"exam_category":       info.ExamCategory,
					"exam_child_category": info.ExamChildCategory,
					"desc":                info.Desc,
				}
				if param.IsRelevanceQuestion {
					updateContent["question_id"] = info.QuestionID
					delete(updateContent, "jiangjie_key")
					delete(updateContent, "question_category")
				}
				bulkSlice = append(bulkSlice, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": sf.ObjectID(info.Id)}).SetUpdate(bson.M{"$set": updateContent}))
			} else {
				// 新增打点信息
				if param.IsRelevanceQuestion {
					var vk models.VideoQuestions
					vk.Id = primitive.NewObjectID()
					vk.CreatedTime = time.Now().Format("2006-01-02 15:04:05")
					vk.UpdatedTime = vk.CreatedTime
					vk.MediaId = param.MediaId
					vk.LessonId = param.LessonId
					vk.ExamCategory = info.ExamCategory
					vk.ExamChildCategory = info.ExamChildCategory
					//vk.QuestionCategory = info.QuestionKeypoints
					//vk.JiangjieKey = info.JiangjieKey
					vk.StartTime = info.StartTime
					vk.EndTime = info.EndTime
					vk.Desc = info.Desc
					vk.QuestionID = info.QuestionID
					updateFilter := bson.M{
						"lesson_id":           param.LessonId,
						"media_id":            param.MediaId,
						"exam_category":       info.ExamCategory,
						"exam_child_category": info.ExamChildCategory,
						"start_time":          info.StartTime,
						"end_time":            info.EndTime,
						"desc":                info.Desc,
					}
					updateFilter["question_id"] = info.QuestionID
					bulkSlice = append(bulkSlice, mongo.NewUpdateOneModel().SetFilter(updateFilter).SetUpdate(bson.M{"$setOnInsert": vk}).SetUpsert(true))
				} else {
					var vk models.VideoKeypoints
					vk.Id = primitive.NewObjectID()
					vk.CreatedTime = time.Now().Format("2006-01-02 15:04:05")
					vk.UpdatedTime = vk.CreatedTime
					vk.MediaId = param.MediaId
					vk.LessonId = param.LessonId
					vk.ExamCategory = info.ExamCategory
					vk.ExamChildCategory = info.ExamChildCategory
					vk.QuestionCategory = info.QuestionKeypoints
					vk.JiangjieKey = info.JiangjieKey
					vk.StartTime = info.StartTime
					vk.EndTime = info.EndTime
					vk.Desc = info.Desc
					updateFilter := bson.M{
						"lesson_id":           param.LessonId,
						"media_id":            param.MediaId,
						"exam_category":       info.ExamCategory,
						"exam_child_category": info.ExamChildCategory,
						"question_category":   info.QuestionKeypoints,
						"start_time":          info.StartTime,
						"end_time":            info.EndTime,
						"desc":                info.Desc,
						"jiangjie_key":        info.JiangjieKey,
					}
					bulkSlice = append(bulkSlice, mongo.NewUpdateOneModel().SetFilter(updateFilter).SetUpdate(bson.M{"$setOnInsert": vk}).SetUpsert(true))
				}
			}

		}
	}
	if len(bulkSlice) > 0 {
		_, err = sf.DB().Collection(tableName).BulkWrite(bulkSlice)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	if len(vkRelevanceBulkSlice) > 0 {
		_, err = sf.DB().Collection("video_keypoints_with_relevancy").BulkWrite(bulkSlice)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	sf.Success(nil, c)
}

func (sf *VideoKeypoint) GetVideoKeypoints(c *gin.Context) {
	var param struct {
		LessonId            int  `json:"lesson_id"`
		IsRelevanceQuestion bool `json:"is_relevance_question"` // 是否是关联试题的打点
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	if param.IsRelevanceQuestion {
		var v []models.VideoQuestions
		err = sf.DB().Collection("video_questions").Where(bson.M{"lesson_id": param.LessonId}).Sort("start_time").Find(&v)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		sf.Success(map[string]interface{}{"return_info": v}, c)
	} else {
		var v []models.VideoKeypoints
		err = sf.DB().Collection("video_keypoints").Where(bson.M{"lesson_id": param.LessonId}).Sort("start_time").Find(&v)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeInvalidParam, c)
			return
		}
		sf.SLogger().Info(v)
		sf.Success(map[string][]models.VideoKeypoints{"return_info": v}, c)
	}
}

func (sf *VideoKeypoint) GetVideoFromKeypoints(c *gin.Context) {
	var param struct {
		ExamCategory         string `json:"exam_category"`
		ExamChildCategory    string `json:"exam_child_category"`
		QuestionKeyPointsStr string `json:"question_category_str"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}

	// 查询表中的考点
	relationVksMap := make(map[string][]models.VideoKeypoints)
	// match := bson.M{"video_url": bson.M{"$exists": true, "$ne": ""}}
	match := bson.M{}
	if param.ExamCategory != "" {
		match["exam_category"] = param.ExamCategory
	}
	if param.ExamChildCategory != "" {
		match["exam_child_category"] = param.ExamChildCategory
	}
	if param.QuestionKeyPointsStr != "" {
		for index, k := range strings.Split(param.QuestionKeyPointsStr, "/") {
			match[fmt.Sprintf("question_category.%d", index)] = k
		}
	}
	aggregate := bson.A{
		bson.M{"$match": match},
		bson.M{"$sort": bson.M{"start_time": 1}},
		bson.M{
			"$group": bson.M{"_id": "$question_category", "video_infos": bson.M{"$push": "$$ROOT"}},
		},
	}
	// 接收聚合查询后的结构体
	type tempData struct {
		ID         []string                `bson:"_id"`
		VideoInfos []models.VideoKeypoints `bson:"video_infos"`
	}
	tempDatas := make([]tempData, 0)
	err = sf.DB().Collection("video_keypoints").Aggregate(aggregate, &tempDatas)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	tempDatasWithRelevancy := make([]tempData, 0)
	err = sf.DB().Collection("video_keypoints_with_relevancy").Aggregate(aggregate, &tempDatasWithRelevancy)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	// 组装考点和视频信息的字典
	for _, t := range tempDatas {
		title := strings.Join(t.ID, "/")
		videoInfos := t.VideoInfos
		relationVksMap[title] = videoInfos
	}

	// 组装考点和视频信息的字典
	for _, t := range tempDatasWithRelevancy {
		title := strings.Join(t.ID, "/")
		videoInfos := t.VideoInfos
		for i, _ := range videoInfos {
			videoInfos[i].IsFromChild = true
		}
		relationVksMap[title] = append(relationVksMap[title], videoInfos...)
	}

	var qcs = new(models.QuestionCategory)
	_ = sf.DB().Collection("question_category").Where(bson.M{"exam_category": param.ExamCategory, "exam_child_category": param.ExamChildCategory}).Take(qcs)
	items := make([]models.QuestionCategoryItem, 0)
	if param.QuestionKeyPointsStr == "" {
		tempItems := qcs.Categorys
		for _, item := range tempItems {
			sf.AddVideoInfoToKeypoints(item.Title, relationVksMap, &item)
			items = append(items, item)
		}
	} else {
		keypointsIndex := make([]int, 0)
		preTitle := make([]string, 0)
		preIndex := make([]int, 0)
		sf.findKeypoints(qcs.Categorys, preTitle, param.QuestionKeyPointsStr, preIndex, &keypointsIndex)
		item := new(models.QuestionCategoryItem)
		for _, index := range keypointsIndex {
			if item.Title == "" {
				item = &qcs.Categorys[index]
			} else {
				item = &item.ChildCategory[index]
			}
		}
		sf.AddVideoInfoToKeypoints(param.QuestionKeyPointsStr, relationVksMap, item)
		items = append(items, *item)
	}
	sf.Success(map[string]interface{}{"return_info": items}, c)

}

func (sf *VideoKeypoint) AddVideoInfoToKeypoints(prefix string, tempMap map[string][]models.VideoKeypoints, qc *models.QuestionCategoryItem) {
	if vks, ok := tempMap[prefix]; ok {
		qc.RelationVideo = vks
	}
	if len(qc.ChildCategory) > 0 {
		for i, _ := range qc.ChildCategory {
			sf.AddVideoInfoToKeypoints(prefix+"/"+qc.ChildCategory[i].Title, tempMap, &qc.ChildCategory[i])
		}
	}
}

func (sf *VideoKeypoint) findKeypoints(arr []models.QuestionCategoryItem, preTitle []string, needleTitle string, preIndex []int, keypointsIndexs *[]int) {
	for i, item := range arr {
		index := preIndex
		text := preTitle
		index = append(index, i)
		text = append(text, item.Title)
		if needleTitle == strings.Join(text, "/") {
			*keypointsIndexs = index
			return
		} else {
			sf.findKeypoints(item.ChildCategory, text, needleTitle, index, keypointsIndexs)
		}
	}
}

func (sf *VideoKeypoint) SaveUpAndDownRelevance(c *gin.Context) {
	var param struct {
		ExamCategory      string   `json:"exam_category"`
		ExamChildCategory string   `json:"exam_child_category"`
		ParentName        []string `json:"parent_name"`
		Ids               []string `json:"ids"`
		Action            string   `json:"action"`
	}
	err := c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	var bulkSlice []mongo.WriteModel

	if param.Action == "delete" {
		bulkSlice = append(bulkSlice, mongo.NewDeleteManyModel().SetFilter(bson.M{"exam_category": param.ExamCategory, "exam_child_category": param.ExamChildCategory, "question_category": param.ParentName, "vk_id": bson.M{"$in": param.Ids}}))
	} else if param.Action == "add" {
		var vks []models.VideoKeypoints
		err = sf.DB().Collection("video_keypoints").Where(bson.M{"_id": bson.M{"$in": sf.ObjectIDs(param.Ids)}}).Find(&vks)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
		for _, vkInfo := range vks {
			var vk models.VideoKeypointsForRelevance
			vk.Id = primitive.NewObjectID()
			vk.VKId = vkInfo.Id.Hex()
			vk.CreatedTime = time.Now().Format("2006-01-02 15:04:05")
			vk.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
			vk.MediaId = vkInfo.MediaId
			vk.LessonId = vkInfo.LessonId
			vk.ExamCategory = param.ExamCategory
			vk.ExamChildCategory = param.ExamChildCategory
			vk.QuestionCategory = param.ParentName
			vk.JiangjieKey = vkInfo.JiangjieKey
			vk.StartTime = vkInfo.StartTime
			vk.EndTime = vkInfo.EndTime
			vk.Desc = vkInfo.Desc
			updateFilter := bson.M{
				"lesson_id":           vkInfo.LessonId,
				"media_id":            vkInfo.MediaId,
				"exam_category":       param.ExamCategory,
				"exam_child_category": param.ExamChildCategory,
				"question_category":   param.ParentName,
				"start_time":          vkInfo.StartTime,
				"end_time":            vkInfo.EndTime,
				"desc":                vkInfo.Desc,
				"jiangjie_key":        vkInfo.JiangjieKey,
			}
			bulkSlice = append(bulkSlice, mongo.NewUpdateOneModel().SetFilter(updateFilter).SetUpdate(bson.M{"$setOnInsert": vk}).SetUpsert(true))
		}
	}

	if len(bulkSlice) > 0 {
		_, err = sf.DB().Collection("video_keypoints_with_relevancy").BulkWrite(bulkSlice)
		if err != nil {
			sf.SLogger().Error(err)
			sf.Error(common.CodeServerBusy, c)
			return
		}
	}
	sf.Success(nil, c)
}
