package jobs

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common/global"
	"interview/grpc/client"
	"interview/models"
	"time"

	"github.com/remeh/sizedwaitgroup"
)

func init() {
	gctt := new(SpeechText)
	RegisterOnceJob(gctt)
}

type SpeechText struct {
	JobBase
}

func (sf *SpeechText) Do() {
	sf.SLogger().Info("启动答题记录和点评语音转写文字服务")
	sw := sizedwaitgroup.New(10)
	callbackUrl := "https://api.xtjzx.cn/interview/app/n/v1/g/speechtext/callback"
	if global.CONFIG.Env == "dev" {
		callbackUrl = "https://dev-api.xtjzx.cn/interview/app/n/v1/g/speechtext/callback"
	}
	var limit int64 = 300
	for {
		filterQuery := bson.M{"speech_text_task_id": "", "created_time": bson.M{"$gt": "2024-04-25 00:00:00"}, "log_type": 2, "answer": bson.M{"$elemMatch": bson.M{"voice_url": bson.M{"$regex": "^https"}, "voice_text": "", "status": bson.M{"$in": []int{0, 1}}}}}
		var logs = make([]models.GAnswerLog, 0)
		sf.DB().Collection("g_interview_answer_logs").Where(filterQuery).Limit(limit).Find(&logs)

		logCount := len(logs)
		for _, log := range logs {
			sw.Add()
			go func(log models.GAnswerLog) {
				defer sw.Done()
				taskid, err := client.CreateRecTask(log.Answer[0].VoiceUrl, callbackUrl)
				if err != nil {
					sf.SLogger().Error("语音转写文字服务 error:", err)
					return
				}
				sf.DB().Collection("g_interview_answer_logs").Where(bson.M{"_id": log.Id}).Update(bson.M{"speech_text_task_id": fmt.Sprintf("%d", taskid), "updated_time": time.Now().Format("2006-01-02 15:04:05")})
			}(log)
		}
		sw.Wait()

		var commentLogList = make([]models.AnswerComment, 0)
		filterQuery = bson.M{"speech_text_task_id": "", "created_time": bson.M{"$gt": "2024-04-25 00:00:00"}, "comment.voice_url": bson.M{"$regex": "^https"}, "comment.voice_text": "", "comment.status": bson.M{"$in": []int{0, 1}}}
		_ = sf.DB().Collection("interview_comment_logs").Where(filterQuery).Limit(limit).Find(&commentLogList)
		commentCount := len(commentLogList)
		for _, log := range commentLogList {
			sw.Add()
			go func(log models.AnswerComment) {
				defer sw.Done()
				taskid, err := client.CreateRecTask(log.Comment.VoiceUrl, callbackUrl)
				if err != nil {
					sf.SLogger().Error("语音转写文字服务 error:", err)
					return
				}
				sf.DB().Collection("interview_comment_logs").Where(bson.M{"_id": log.Id}).Update(bson.M{"speech_text_task_id": fmt.Sprintf("%d", taskid), "updated_time": time.Now().Format("2006-01-02 15:04:05")})
			}(log)
		}
		sf.SLogger().Info("语音转写文字服务 log count:", logCount, " comment count:", commentCount)
		if int64(logCount) < limit && int64(commentCount) < limit {
			time.Sleep(1 * time.Minute)
		}
	}
}
