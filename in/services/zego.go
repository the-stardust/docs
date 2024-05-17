package services

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zegoim/zego_server_assistant/token/go/src/token04"
	"interview/common"
	"interview/common/global"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Zego struct {
	ServicesBase
}

var ZegoBaseUrl = "https://cloudrecord-api.zego.im/?Action="
var zegoAppId string = global.CONFIG.Zego.AppId
var zegoServerSecret string = global.CONFIG.Zego.Secret

func NewZego() *Zego {
	return new(Zego)
}

// https://doc-zh.zego.im/article/7646#3
func (sf *Zego) GetPrivilegeToken(roomId, userId string) (string, error) {
	var err error
	var token string
	appIdUint64, _ := strconv.ParseUint(zegoAppId, 10, 64)
	var appId uint32 = uint32(appIdUint64)
	var serverSecret = zegoServerSecret
	var effectiveTimeInSeconds int64 = 7200 // token 的有效时长，单位：秒

	////权限位定义
	//const (
	//	PrivilegeKeyLogin   = 1 // 1 代表登录权限
	//	PrivilegeKeyPublish = 2 // 2 代表推流权限
	//)

	////权限开关定义
	//const (
	//	PrivilegeEnable     = 1 // 允许相应业务权限
	//	PrivilegeDisable    = 0 // 不允许相应业务权限
	//)
	//privilege := make(map[int]int)
	//privilege[token04.PrivilegeKeyLogin] = token04.PrivilegeEnable
	//privilege[token04.PrivilegeKeyPublish] = token04.PrivilegeEnable
	////token业务扩展：权限认证属性
	//type RtcRoomPayLoad struct {
	//	RoomId       string      `json:"room_id"`        //房间id；用于对接口的房间id进行强验证
	//	Privilege    map[int]int `json:"privilege"`      //权限位开关列表；用于对接口的操作权限进行强验证
	//	StreamIdList []string    `json:"stream_id_list"` //流列表；用于对接口的流id进行强验证；允许为空，如果为空，则不对流id验证
	//}
	////token业务扩展配置
	//payloadData := &RtcRoomPayLoad{
	//	RoomId:       roomId,
	//	Privilege:    privilege,
	//	StreamIdList: nil,
	//}
	//
	//payload, err := json.Marshal(payloadData)
	//if err != nil {
	//	sf.SLogger().Error(err)
	//	return token, err
	//}

	payload := ""

	//生成token
	token, err = token04.GenerateToken04(appId, userId, serverSecret, effectiveTimeInSeconds, string(payload))
	if err != nil {
		sf.SLogger().Error(err)
	}
	return token, err
}

// php示例：
// $signature = $_POST["signature"];
// $timestamp = $_POST["timestamp"];
// $nonce = $_POST["nonce"];
//
// $secret = callbacksecret;//后台获取的callbacksecret
// $tmpArr = array($secret, $timestamp, $nonce);
// sort($tmpArr, SORT_STRING);
// $tmpStr = implode( $tmpArr );
// $tmpStr = sha1( $tmpStr );
//
// if( $tmpStr == $signature ){
// return true;
// } else {
// return false;
// }
func (sf *Zego) CheckSign(Signature, Nonce string, Timestamp string) bool {
	callbackSecret := global.CONFIG.Zego.CallbackSecret //"ab6a0c84eaabc3e5798c52a299b3fec9" //"8765f027f2353d0e32e25b85afcc6aeb"
	tmpArr := []string{callbackSecret, Timestamp, Nonce}
	sort.Slice(tmpArr, func(i, j int) bool {
		return tmpArr[i] < tmpArr[j]
	})

	h := sha1.New()
	io.WriteString(h, strings.Join(tmpArr, ""))
	return Signature == fmt.Sprintf("%x", h.Sum(nil))
}

func (sf *Zego) getUrl(action string) string {
	nonceByte := make([]byte, 8)
	rand.Read(nonceByte)
	nonce := hex.EncodeToString(nonceByte)
	timestamp := time.Now().Unix()
	signature := sf.GenerateSignature(nonce, timestamp)
	return fmt.Sprintf(
		"%s&AppId=%s&SignatureNonce=%s&Timestamp=%d&Signature=%s&SignatureVersion=2.0&IsTest=false",
		ZegoBaseUrl+action,
		zegoAppId,
		nonce,
		timestamp,
		signature,
	)
}

// https://doc-zh.zego.im/article/12322
func (sf *Zego) StartVideoRecord(roomId, uid string) (string, error) {
	url := sf.getUrl("StartRecord")
	param := map[string]any{
		"RoomId": roomId,
		"RecordInputParams": map[string]any{
			//"RecordMode":  1,
			//"StreamType":  3,
			//"MaxIdleTime": 60,
			//"RecordStreamList": []map[string]string{
			//	{
			//		"StreamId": uid,
			//	},
			//},

			"RecordMode": 2,
			"StreamType": 3,
			"FillBlank":  true,
			"FillFrame": map[string]any{
				"FrameFillMode":  2,
				"FrameFillColor": 4290167040,
			},
			"MaxIdleTime": 120,
			"MixConfig": map[string]any{
				"MixMode":           2,
				"MixOutputStreamId": uid,
				"MixOutputVideoConfig": map[string]any{
					"Width":   320,
					"Height":  240,
					"Fps":     15,
					"Bitrate": 200000,
				},
				"MixOutputAudioConfig": map[string]any{
					"Bitrate": 48000,
				},
			},
		},
		"RecordOutputParams": map[string]any{
			"OutputFileFormat": "mp4",
			"OutputFolder":     roomId + "/",
		},
		"StorageParams": map[string]any{
			"Vendor":          2,
			"Region":          "oss-cn-zhangjiakou",
			"Bucket":          "xtj-interview-mock",
			"AccessKeyId":     global.CONFIG.MockExamOSS.AppKeyId,
			"AccessKeySecret": global.CONFIG.MockExamOSS.Secret,
		},
	}
	dataByte, err := common.HttpPostJson(url, param)
	if err != nil {
		sf.SLogger().Error(err)
		return "", err
	}
	type respTaskId struct {
		TaskId string `json:"TaskId"`
	}
	type resp struct {
		Code      int64      `json:"Code"`
		Message   string     `json:"Message"`
		RequestId string     `json:"RequestId"`
		Data      respTaskId `json:"Data"`
	}
	r := new(resp)
	err = json.Unmarshal(dataByte, r)
	if err != nil {
		sf.SLogger().Error(err, " resp:"+string(dataByte))
		return "", err
	}
	if r.Code != 0 {
		sf.SLogger().Error(r.Message, " roomId:"+string(roomId))
		return "", errors.New(r.Message)
	}

	return r.Data.TaskId, nil
}

// https://doc-zh.zego.im/article/12350
func (sf *Zego) EndVideoRecord(taskId string) error {
	url := sf.getUrl("StopRecord")
	param := map[string]any{
		"TaskId": taskId,
	}
	dataByte, err := common.HttpPostJson(url, param)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	type resp struct {
		Code      int64  `json:"Code"`
		Message   string `json:"Message"`
		RequestId string `json:"RequestId"`
	}
	r := new(resp)
	err = json.Unmarshal(dataByte, r)
	if err != nil {
		sf.SLogger().Error(err, "resp:"+string(dataByte))
		return err
	}
	if r.Code != 0 {
		sf.SLogger().Error(r.Message, " taskId:"+string(taskId))
		return errors.New(r.Message)
	}
	return nil
}

// Signature=md5(AppId + SignatureNonce + ServerSecret + Timestamp)
func (sf *Zego) GenerateSignature(signatureNonce string, timestamp int64) (Signature string) {
	data := fmt.Sprintf("%s%s%s%d", zegoAppId, signatureNonce, zegoServerSecret, timestamp)
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
