package services

import (
	"encoding/json"
	"fmt"
	"interview/common"
	"interview/common/global"
)

func GetQuestionVideos(qid, cateStr string) (interface{}, interface{}, error) {
	if qid == "" && cateStr == "" {
		return nil, nil, fmt.Errorf("qid and cate str empty")
	}
	params := map[string]interface{}{
		"question_id":  qid,
		"jiangjie_key": cateStr,
	}
	strParams := common.HttpBuildQuery(params)
	url := global.CONFIG.ServiceUrls.QuestionBankUrl + "/question-bank/app/n/v1/question/videos?" + strParams

	respData, err := common.HttpGet(url)
	if err != nil {
		return nil, nil, err
	}
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			CategoryVideoList interface{} `json:"category_video_list"`
			QuestionVideoList interface{} `json:"question_video_list"`
		} `json:"data"`
	}

	_ = json.Unmarshal(respData, &res)
	if res.Code > 0 {
		return nil, nil, fmt.Errorf("%s", res.Message)
	}
	return res.Data.CategoryVideoList, res.Data.QuestionVideoList, nil
}
