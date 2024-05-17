package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/helper"
	"strings"

	"github.com/segmentio/kafka-go"
)

func init() {
	s := new(GPTPreview)
	RegisterOnceJob(s)
}

type GPTPreview struct {
	JobBase
}

// msg struct
type CommonResponse struct {
	Content []string          `json:"content"`
	Extra   map[string]string `json:"extra"` // to_topic需要用到的一些额外信息
}

func (sf *GPTPreview) Do() {
	sf.SLogger().Info("启动gpt预生成消费队列")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(global.CONFIG.Kafka.Address, ","),
		GroupID:  global.CONFIG.Kafka.GPTServiceGroupId,
		Topic:    global.CONFIG.Kafka.GPTQuestionPreviewTopic,
		MaxBytes: 10e6, // 10MB
	})

	ctx := context.Background()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			sf.SLogger().Error(err)
			break
		}

		//用户答题统计
		var resp = new(CommonResponse)
		err = json.Unmarshal(m.Value, resp)
		if err != nil {
			sf.SLogger().Error(err)
			continue
		}

		extra := resp.Extra
		questionId := extra["question_id"]
		previewType := extra["preview_type"]
		if from, ok := extra["from"]; !ok || from != "interview" {
			if err := r.CommitMessages(ctx, m); err != nil {
				sf.SLogger().Error("failed to commit messages:", err)
			}
			continue
		}

		cacheKey := fmt.Sprintf("%s%s", rediskey.GPTQuestionPreview, questionId)
		str, err := helper.RedisGet(cacheKey)

		previewItem := make(map[string][]string, 0)
		if str != "" {
			err = json.Unmarshal([]byte(str), &previewItem)
			if err != nil {
				sf.SLogger().Error(err)
				continue
			}
		}
		if _, ok := previewItem[previewType]; !ok {
			previewItem[previewType] = make([]string, 0)
		}
		previewItem[previewType] = append(previewItem[previewType], resp.Content[0])
		piByte, _ := json.Marshal(previewItem)
		err = helper.RedisSet(cacheKey, string(piByte), 7*86400)
		// 标准答案和答题思路
		if len(previewItem) > 1 {
			previewingStatus := getPreviewStatus(previewItem)
			helper.RedisSet(fmt.Sprintf("%s%s", rediskey.GPTQuestionPreviewing, questionId), previewingStatus, 6*86400)
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			sf.SLogger().Error("failed to commit messages:", err)
		}
	}
	if err := r.Close(); err != nil {
		sf.SLogger().Error(err)
	} else {
		sf.SLogger().Info("gpt预生成消费队列关闭")
	}
}

func getPreviewStatus(previewItem map[string][]string) string {
	// 默认失败
	// 如果有内容不是 出错了，暂时无法生成 就认为是部分生成
	// 如果得分点和答题思路都有 并且可以被解析 就认为是已生成
	previewingStatus := "4"
	previewingStatus3 := false
	for _, prei := range previewItem {
		for _, cont := range prei {
			if cont != "出错了，暂时无法生成" {
				previewingStatus3 = true
				break
			}
		}
		if previewingStatus3 {
			previewingStatus = "3"
			break
		}
	}
	if previewingStatus == "3" {
		var previewItemLen = 0
		for _, val := range previewItem {
			for _, v := range val {
				if v != "出错了，暂时无法生成" {
					previewItemLen++
					break
				}
			}
		}
		if len(previewItem) >= previewItemLen {
			previewingStatus = "2"
		}
	}

	return previewingStatus
}
