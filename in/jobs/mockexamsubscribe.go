package jobs

import (
	"fmt"
	"github.com/remeh/sizedwaitgroup"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common/rediskey"
	"interview/helper"
	"interview/models"
	"interview/services"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	questionCountJob := new(MockExamSubscribe)
	RegisterWorkConnectJob(questionCountJob, 120, 1)
}

type MockExamSubscribe struct {
	JobBase
}

func (sf *MockExamSubscribe) GetJobs() ([]interface{}, error) {
	var err error
	var nowStr = time.Now().Add(20 * time.Minute).Format("2006-01-02 15:04:05") // 20分钟后的时间
	var qc = make([]models.MockExam, 0)
	var filter = bson.M{
		"exam_info.start_time": bson.M{"$lt": nowStr},
		"exam_info.end_time":   bson.M{"$gt": nowStr},
		"status":               1,
		"sign_up_user_num":     bson.M{"$gt": 0},
	}
	sf.SLogger().Info("启动模考订阅通知任务 GetJobs: " + nowStr)
	err = sf.DB().Collection("mock_exam").Where(filter).Find(&qc)
	if err != nil {
		return []interface{}{}, err
	}
	var ids []interface{}
	for _, v := range qc {
		ids = append(ids, v.Id.Hex())
	}
	return ids, nil
}

func (sf *MockExamSubscribe) Do(mockExamIDAny interface{}) error {
	mockExamID := mockExamIDAny.(string)
	sf.SLogger().Info("启动模考订阅通知任务 Do")
	var timeDateStr = time.Now().Format("2006-01-02")
	var time20 = time.Now().Add(20 * time.Minute) // 20分钟后的时间
	var time20TimeStr = time20.Format("15:04:05") // 20分钟后的时间
	var err error
	var exam = new(models.MockExam)
	err = sf.DB().Collection("mock_exam").Where(bson.M{"_id": sf.ObjectID(mockExamID)}).Take(exam)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	var qc = make([]models.MockExamLog, 0)
	err = sf.DB().Collection("mock_exam_log").Where(bson.M{
		"mock_exam_id": mockExamID, "slot_date": timeDateStr, "slot_start_time": bson.M{"$lt": time20TimeStr},
		"subscribe_status": 1, "status": bson.M{"$ne": -1},
	}).Find(&qc)
	if err != nil && !sf.MongoNoResult(err) {
		sf.SLogger().Error(err)
		return err
	}
	if len(qc) == 0 {
		return nil
	}
	templItem := make(map[string]services.TemplItem)
	key := fmt.Sprintf("%s%s", string(rediskey.MockExamSubscribe), mockExamID)
	var ids = make([]primitive.ObjectID, 0)
	swg := sizedwaitgroup.New(20)
	for _, log := range qc {
		openId, err := helper.RedisHGet(key, log.UserId)
		if err != nil || openId == "" {
			continue
		}
		swg.Add()
		templItem["thing1"] = services.TemplItem{
			Value: exam.Title,
		}
		templItem["time3"] = services.TemplItem{
			Value: log.SlotStartTime,
		}
		ids = append(ids, log.Id)
		msgReq := services.SubscribeMsgReq{
			Touser:    openId,
			TemplID:   services.GetWeChatMockExamTemplateId(),
			Page:      "/mock/mockInfo/index?exam_id=" + mockExamID,
			TemplData: templItem,
		}
		go func(msgReq services.SubscribeMsgReq) {
			defer swg.Done()
			err = services.SendSubscribeMsg(msgReq)
			if err != nil {
				sf.SLogger().Error(err)
			}
		}(msgReq)
	}
	swg.Wait()
	sf.DB().Collection("mock_exam_log").Where(bson.M{"_id": bson.M{"$in": ids}}).Update(bson.M{"subscribe_status": 2})
	return nil
}
