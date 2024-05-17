package qb_crm_data

import (
	"encoding/json"
	"errors"
	"interview/common"
	"interview/common/global"
	"interview/services"
)

type kqRecordSrv struct {
	services.ServicesBase
}

var (
	KqRecordSrv = &kqRecordSrv{}
)

type DataRespV2 struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func getFullUrl(reqPath string) string {
	return global.CONFIG.ServiceUrls.CrmDataServiceUrl + reqPath
}

func (this *kqRecordSrv) GetUserClassName(guid string) ([]string, error) {
	respData, err := common.HttpGet(getFullUrl("/question-bank-crm-data/inner/kqRecord/getUserClassName?uid=" + guid))
	if err != nil {
		return nil, err
	}

	infoList := make([]string, 0)
	resp := &DataRespV2{Data: &infoList}
	_ = json.Unmarshal(respData, resp)
	if resp.Code > 0 {
		return nil, errors.New(resp.Message)
	}

	return infoList, nil
}
