package services

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/helper"
	"io"
	"net/http"
	"sort"
	"sync"
)

var appid = global.CONFIG.Wechat.AppId
var secret = global.CONFIG.Wechat.Secret
var token = "7bfd32c894538372"
var mlock = sync.Mutex{}

func GetWeChatMockExamTemplateId() string {
	return "IoEaxO4A2Yh9VZSWpkFaYRzd8Ouys4LdNvcuHqSvsLQ"
}

func GetAccessToken() (string, error) {
	if helper.WxAccessTokenTTL() < 100 {
		mlock.Lock()
		defer mlock.Unlock()
		ret, err := getAccessToken()
		if err != nil {
			return "", err
		}
		return ret.AccessToken, nil
	}
	return helper.RedisGet(string(rediskey.WxAccessToken))
}

type WxMsg struct {
	ToUserName   string    `json:"ToUserName"`
	FromUserName string    `json:"FromUserName"`
	CreateTime   int       `json:"CreateTime"`
	MsgType      string    `json:"MsgType"`
	Event        string    `json:"Event"`
	List         WxMsgList `json:"List"`
}

type WxMsgList struct {
	PopupScene            int    `json:"PopupScene"`
	SubscribeStatusString string `json:"SubscribeStatusString"`
	TemplateId            string `json:"TemplateId"`
}

type WxApiRet struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type WxApiTokenRet struct {
	WxApiRet
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

/**
*{
*   "errcode": 0,
*   "errmsg": "ok",
*   "data": [
*       {
*          "priTmplId": "9Aw5ZV1j9xdWTFEkqCpZ7mIBbSC34khK55OtzUPl0rU",
*          "title": "报名结果通知",
*          "content": "会议时间:{{date2.DATA}}\n会议地点:{{thing1.DATA}}\n",
*          "example": "会议时间:2016年8月8日\n会议地点:TIT会议室\n",
*          "type": 2
*       },
*       {
*          "priTmplId": "cy_DfOZL7lypxHh3ja3DyAUbn1GYQRGwezuy5LBTFME",
*          "title": "洗衣机故障提醒",
*          "content": "完成时间:{{time1.DATA}}\n所在位置:{{enum_string2.DATA}}\n提示说明:{{enum_string3.DATA}}\n",
*          "example": "完成时间:2021年10月21日 12:00:00\n所在位置:客厅\n提示说明:设备发生故障，导致工作异常，请及时查看\n",
*          "keywordEnumValueList": [
*             {
*                "enumValueList": ["客厅","餐厅","厨房","卧室","主卧","次卧","客卧","父母房","儿童房","男孩房","女孩房","卫生间","主卧卫生间","公共卫生间","衣帽间","书房","游戏室","阳台","地下室","储物间","车库","保姆房","其他房间"],
*                "keywordCode": "enum_string2.DATA"
*             },
*             {
*                "enumValueList": ["设备发生故障，导致工作异常，请及时查看"],
*                "keywordCode": "enum_string3.DATA"
*             }
*          ],
*          "type": 3
*       }
*   ]
*}
 */
type WxTemplateRet struct {
	WxApiRet
	Data []WxTemplate `json:"data"`
}

type WxTemplate struct {
	PriTmplId            string                 `json:"priTmplId"`
	Title                string                 `json:"title"`
	Content              string                 `json:"content"`
	Example              string                 `json:"example"`
	Type                 int                    `json:"type"` // 模版类型，2 为一次性订阅，3 为长期订阅
	KeywordEnumValueList []KeywordEnumValueList `json:"keywordEnumValueList"`
}

type KeywordEnumValueList struct {
	KeywordCode   string   `json:"keywordCode"`
	EnumValueList []string `json:"enumValueList"`
}

//	{
//		"template_id": "lHRdQMIhHpxx184-xT07nCQh2oozCuzv0V71oNKx8pU",
//		"title": "题库功能提醒",
//		"example": "科目:执业药师-药学专业知识（一））\n题库功能:每日一练刷题做题通知\n备注:点击卡片，前往做题>\n",
//		"content": [
//			{
//				"key": "thing1",
//				"value": "",
//				"content": "科目:"
//			},
//			{
//				"key": "thing2",
//				"value": "",
//				"content": "题库功能:"
//			},
//			{
//				"key": "thing3",
//				"value": "",
//				"content": "备注:"
//			}
//		]
//	}
type RespItem struct {
	TemplateId string        `json:"template_id"`
	Title      string        `json:"title"`
	Example    string        `json:"example"`
	Page       string        `json:"page"`
	Content    []ItemContent `json:"content"`
}

// 资料、 合集 API所需结构
type ItemContent struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Content string `json:"content"`
	Example string `json:"example"`
}

type WxSessionData struct {
	WxApiRet
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	Sessionkey string `json:"session_key"`
}

// 获取小程序全局唯一后台接口调用凭据（access_token）。
// 调用绝大多数后台接口时都需使用 access_token，开发者需要进行妥善保存。
func getAccessToken() (WxApiTokenRet, error) {
	wx_api_url := "https://api.weixin.qq.com/cgi-bin/token"
	wx_api_url = fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s",
		wx_api_url, appid, secret)

	var ent WxApiTokenRet
	res, err := WxApiGet(wx_api_url)
	if err != nil {
		return ent, err
	}

	if err := json.Unmarshal(res, &ent); err != nil {
		return ent, err
	}

	if ent.ErrCode != 0 {
		return ent, errors.New(ent.ErrMsg)
	}
	helper.RedisSet(string(rediskey.WxAccessToken), ent.AccessToken, int(ent.ExpiresIn))
	return ent, nil
}

func WxApiGet(wx_addr string) ([]byte, error) {
	res, err := http.Get(wx_addr)
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http statusCode=%v", res.StatusCode))
	}

	return raw, nil
}

func WxApiPost(wx_addr string, data []byte) ([]byte, error) {
	res, err := http.Post(wx_addr, "application/json; charset=utf-8", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http statusCode=%v", res.StatusCode))
	}

	return raw, nil
}

// 消息推送接入请求验证
func WeixinCheck(timestamp, nonce, signature string) bool {
	tmps := []string{token, timestamp, nonce}
	sort.Strings(tmps)
	tmpStr := tmps[0] + tmps[1] + tmps[2]
	tmp := str2sha1(tmpStr)
	if tmp == signature {
		return true
	} else {
		return false
	}
}

func str2sha1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

type TemplData map[string]TemplItem
type SubscribeMsgReq struct {
	//接收者（用户）的 openid
	Touser string `json:"touser"`
	//所需下发的订阅模板id
	TemplID string `json:"template_id"`
	//点击模板卡片后的跳转页面，仅限本小程序内的页面。支持带参数,（示例index?foo=bar）。该字段不填则模板无跳转。
	Page string `json:"page"`
	//模板内容，格式形如 { "key1": { "value": any }, "key2": { "value": any } }
	TemplData TemplData `json:"data"`
	//跳转小程序类型：developer为开发版；trial为体验版；formal为正式版；默认为正式版
	AppEnv string `json:"miniprogram_state"`
	//支持zh_CN(简体中文)、en_US(英文)、zh_HK(繁体中文)、zh_TW(繁体中文)，默认为zh_CN
	Lang string `json:"lang"`
}

type TemplItem struct {
	Value string `json:"value"`
}

// 发送微信订阅消息
func SendSubscribeMsg(param SubscribeMsgReq, hasRetry ...int) error {
	accessToken, err := GetAccessToken()
	if err != nil {
		return err
	}
	var ent WxApiRet
	wx_addr := "https://api.weixin.qq.com/cgi-bin/message/subscribe/send"
	wx_addr += "?access_token=" + accessToken
	res, err := WxApiPostStruct(wx_addr, param)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(res, &ent); err != nil {
		return errors.New("json fail SendSubscribeMsg")
	}
	if ent.ErrCode == 40001 {
		_, err = getAccessToken()
		if err != nil {
			return err
		}
		// 重试一次
		if len(hasRetry) == 0 {
			return SendSubscribeMsg(param, 1)
		}
	}
	return nil
}

func WxApiPostStruct(wx_addr string, param interface{}) ([]byte, error) {
	data, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	return WxApiPost(wx_addr, data)
}

// docs: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-message-management/subscribe-message/getMessageTemplateList.html
func GetWxTemplates(hasRetry ...int) (WxTemplateRet, error) {
	accessToken, err := GetAccessToken()
	if err != nil {
		return WxTemplateRet{}, err
	}

	wx_api_url := "https://api.weixin.qq.com/wxaapi/newtmpl/gettemplate"
	wx_api_url = fmt.Sprintf("%s?access_token=%s", wx_api_url, accessToken)

	var ent WxTemplateRet
	res, err := WxApiGet(wx_api_url)
	if err != nil {
		return ent, err
	}

	if err := json.Unmarshal(res, &ent); err != nil {
		return ent, err
	}

	if ent.ErrCode != 0 {
		if ent.ErrCode == 40001 {
			_, err = getAccessToken()
			if err != nil {
				return ent, err
			}
			// 重试一次
			if len(hasRetry) == 0 {
				return GetWxTemplates(1)
			}
		}
		return ent, errors.New(ent.ErrMsg)
	}
	return ent, nil
}

// 登录凭证校验。
// 通过 wx.login 接口获得临时登录凭证 code 后传到开发者服务器调用此接口完成登录流程。
// 根据code获取opendid以及session_key
func GetOpenIdByCode(code string) (WxSessionData, error) {
	wx_addr := "https://api.weixin.qq.com/sns/jscode2session"
	wx_addr = fmt.Sprintf("%s?appid=%s&secret=%s&js_code=%s&grant_type=%s",
		wx_addr, appid, secret, code, "authorization_code")

	var ent WxSessionData
	res, err := WxApiGet(wx_addr)
	if err != nil {
		return ent, err
	}

	err = json.Unmarshal(res, &ent)
	if err != nil {
		return ent, err
	}

	return ent, nil
}
