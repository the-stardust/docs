package jobs

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"interview/helper"
	"interview/services"
	"time"
)

func init() {
	s := new(LogStatistics)
	RegisterOnceJob(s)
}

type LogStatistics struct {
	JobBase
}

// 跑昨天的log数据到统计表中
func (sf *LogStatistics) Do() {
	sf.SLogger().Info("启动g_interview_answer_logs统计队列")
	cacheKey := "interview:log_statistics_date"
	date, _ := helper.RedisGet(cacheKey)
	start := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	end := time.Now().Format("2006-01-02")
	if date == start {
		sf.SLogger().Info(fmt.Sprintf("启动g_interview_answer_logs统计队列 %s重复执行", start))
		return
	}
	filter := bson.M{"created_time": bson.M{"$gte": start, "$lt": end}}
	services.NewStatisticsSrv().DailyUserTestLogStatistics2(filter)
	helper.RedisSet(cacheKey, start, 86400*30)
}
