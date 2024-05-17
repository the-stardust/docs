package services

import (
	"context"
	"errors"
	"fmt"
	"interview/common/global"
	"interview/models"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type GPT struct {
	ServicesBase
}
type GPTAnalysis struct {
	Idx     int
	Content []string
}

func (sf *GPT) MakeAnswer(temperature float32, topP float32, systemContent string, prompt string, retryTimes int) ([]string, error) {
	sf.SLogger().Debugf("\ntemperature:%f\ntopP:%f\nsystemContent:%s\nprompt:%s", temperature, topP, systemContent, prompt)
	config := openai.DefaultAzureConfig(global.CONFIG.GPT.ApiKey, global.CONFIG.GPT.BaseUrl)
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			"gpt-3.5-turbo-0301": global.CONFIG.GPT.Engine,
		}
		return azureModelMapping[model]
	}
	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			TopP:        topP,
			Temperature: temperature,
			MaxTokens:   1500,
			Model:       openai.GPT3Dot5Turbo0301,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemContent,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		if retryTimes > 0 {
			return sf.MakeAnswer(temperature, topP, systemContent, prompt, retryTimes-1)
		}
		fmt.Printf("ChatCompletion error: %v\n", err)
		return []string{}, err
	}
	newAnalysis := []string{}
	for _, v := range resp.Choices {
		newAnalysis = append(newAnalysis, v.Message.Content)
	}
	return newAnalysis, err
}
func (sf *GPT) MakeAnswer1(temperature float32, topP float32, chatCompletionMessage []openai.ChatCompletionMessage, retryTimes int, isUse4 bool) ([]string, error) {
	// 默认为GPT3
	config := openai.DefaultAzureConfig(global.CONFIG.GPT.ApiKey, global.CONFIG.GPT.BaseUrl)
	modelMapping := map[string]string{
		"gpt-3.5-turbo-0301": global.CONFIG.GPT.Engine,
	}
	maxTokens := 1500
	model := openai.GPT3Dot5Turbo0301
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := modelMapping
		return azureModelMapping[model]
	}

	// 配置为GPT4
	if isUse4 == true {
		config = openai.DefaultAzureConfig(global.CONFIG.GPT4.ApiKey, global.CONFIG.GPT4.BaseUrl)
		modelMapping = map[string]string{
			"gpt-4": global.CONFIG.GPT4.Engine,
		}
		maxTokens = 1500
		model = openai.GPT4
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := modelMapping
			return azureModelMapping[model]
		}
	}

	// 启动！
	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			TopP:        topP,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			Model:       model,
			Messages:    chatCompletionMessage,
		},
	)

	if err != nil {
		if retryTimes > 0 {
			return sf.MakeAnswer1(temperature, topP, chatCompletionMessage, retryTimes-1, isUse4)
		}
		sf.SLogger().Error(err)
		return []string{}, err
	}
	newAnalysis := []string{}
	for _, v := range resp.Choices {
		newAnalysis = append(newAnalysis, v.Message.Content)
	}
	return newAnalysis, err
}
func (sf *GPT) MakeAnswerStream(temperature float32, topP float32, chatCompletionMessage []openai.ChatCompletionMessage, retryTimes int, cb func(err error, s string, end bool)) {
	ctx := context.Background()
	config := openai.DefaultAzureConfig(global.CONFIG.GPT.ApiKey, global.CONFIG.GPT.BaseUrl)
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			"gpt-3.5-turbo-0301": global.CONFIG.GPT.Engine,
		}
		return azureModelMapping[model]
	}
	client := openai.NewClientWithConfig(config)
	request := openai.ChatCompletionRequest{
		TopP:        topP,
		Temperature: temperature,
		MaxTokens:   1000,
		Model:       openai.GPT3Dot5Turbo0301,
		Messages:    chatCompletionMessage,
		Stream:      false,
	}
	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		sf.SLogger().Error(err)
		cb(err, "", false)
		return
	}
	defer stream.Close()
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			cb(nil, "", true)
			return
		}
		if err != nil {
			sf.SLogger().Error(err)
			cb(err, "", false)
			return
		}
		if len(response.Choices) > 0 {
			cb(nil, response.Choices[0].Delta.Content, false)
		} else {
			sf.SLogger().Debug(response.Choices)
		}
	}

}
func (sf *GPT) MakeAnswerStream1(temperature float32, topP float32, chatCompletionMessage []openai.ChatCompletionMessage, retryTimes int) {
	// sArr, end, err := sf.MakeAnswerStream2(temperature, topP, chatCompletionMessage, 1)
	// sf.SLogger().Debugf("%+v", chatCompletionMessage)
	// if err == nil {
	// 	if !end {

	// 	} else {
	// 		chatCompletionMessage = append(chatCompletionMessage, openai.ChatCompletionMessage{
	// 			Role:    openai.ChatMessageRoleAssistant,
	// 			Content: strings.Join(sArr, ""),
	// 		})
	// 		// chatCompletionMessage = append(chatCompletionMessage, openai.ChatCompletionMessage{
	// 		// 	Role:    openai.ChatMessageRoleUser,
	// 		// 	Content: "继续",
	// 		// })

	// 		// time.Sleep(2 * time.Second)
	// 		// sf.SLogger().Debug("下一句\n")
	// 		// sf.MakeAnswerStream1(temperature, topP, chatCompletionMessage, 1)
	// 	}
	// } else {

	// }
	t := models.GQuestionCategory{
		Name:   "ceshi",
		Prompt: "ceshi",
	}
	sf.DB().Collection("g_interview_question_category").Create(&t)
}
func (sf *GPT) MakeAnswerStream2(temperature float32, topP float32, chatCompletionMessage []openai.ChatCompletionMessage, retryTimes int) (s []string, end bool, err error) {
	ctx := context.Background()
	config := openai.DefaultAzureConfig(global.CONFIG.GPT.ApiKey, global.CONFIG.GPT.BaseUrl)
	config.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			"gpt-3.5-turbo-0301": global.CONFIG.GPT.Engine,
		}
		return azureModelMapping[model]
	}
	client := openai.NewClientWithConfig(config)
	request := openai.ChatCompletionRequest{
		TopP:        topP,
		Temperature: temperature,
		MaxTokens:   1000,
		Model:       openai.GPT3Dot5Turbo0301,
		Messages:    chatCompletionMessage,
		Stream:      true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		sf.SLogger().Error(err)
		return make([]string, 0), false, err
	}
	defer stream.Close()
	resp := []string{}
	var eof = false
	for {
		// time.Sleep(2 * time.Second)
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			eof = true
			break
		}
		if err != nil {
			sf.SLogger().Error(err)
			eof = false
			break
		}
		if len(response.Choices) > 0 {

			resp = append(resp, response.Choices[0].Delta.Content)
		} else {
			sf.SLogger().Debug("gpt返回空")
			break
		}
	}
	sf.SLogger().Debugf("%+v", strings.Join(resp, ""))
	return resp, eof, err
}
