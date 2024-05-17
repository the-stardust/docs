package services

import (
	"encoding/json"
	"interview/common"
	"interview/common/global"
)

type Upload struct {
	ServicesBase
}
type UploadRes struct {
	Status string `json:"status"`
	Data   struct {
		Url string `json:"url"`
	} `json:"data"`
}

// 上传文件
func (sf *Upload) UploadFile(objectName string, fileDownloadName string, filebyte []byte, bucketName ...string) (string, error) {
	var err error
	bucket := "xtj-interview"
	if len(bucketName) > 0 && bucketName[0] != "" {
		bucket = bucketName[0]
	}
	uploadRes, err := common.HttpPostFile(global.CONFIG.ServiceUrls.UploadUrl, map[string]interface{}{"objectName": objectName, "fileDownloadName": fileDownloadName, "bucketName": bucket, "attachment": true, "devToken": "data-wj"}, []common.UploadFile{{FieldName: "file", FileName: "file", Data: filebyte}})
	if err != nil {
		sf.SLogger().Error(err)
		return "", err
	}
	res := UploadRes{}
	err = json.Unmarshal(uploadRes, &res)
	if err != nil {
		sf.SLogger().Error(err)
		return "", err
	}
	return res.Data.Url, err
}
