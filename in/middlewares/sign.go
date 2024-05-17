package middlewares

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"interview/common"
	"io"
	"strconv"
	"time"
)

type Sign struct {
	Middleware
}

// 请求限制：
// 1、 对referer做限制
// 2、对签名做限制
func (sf *Sign) HandlerFunc(c *gin.Context) {
	appId := c.GetHeader("X-XTJ-APPID")
	if appId == "" {
		appId = "wx838f95f8c3c1ea68"
	}
	requestUrl := c.Request.URL
	path := requestUrl.Path
	method := c.Request.Method
	//refer := c.Request.Header.Get("Referer")

	//allowReferList := []string{"https://servicewechat.com", "https://web.xtjzx.cn"}
	//isAllow := false
	//for _, s := range allowReferList {
	//	if strings.Contains(refer, s) {
	//		isAllow = true
	//		break
	//	}
	//}
	//if !isAllow {
	//	c.JSON(403, map[string]any{"code": 403, "msg": "no permission"})
	//	sf.SLogger().Error("fail to get request body: refer:" + refer)
	//	c.Abort()
	//	return
	//}
	timestamp := ""
	sign := ""
	if method != "GET" {
		body, err := c.GetRawData()
		if err != nil {
			sf.SLogger().Error("fail to get request body", err)
		} else {
			requestBody := make(map[string]any)
			d := json.NewDecoder(bytes.NewReader(body))
			d.UseNumber()
			d.Decode(&requestBody)
			if _, ok := requestBody["timestamp"]; ok {
				timestamp = requestBody["timestamp"].(json.Number).String()
				sign, _ = requestBody["sign"].(string)
			}
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	} else {
		timestamp = c.Query("timestamp")
		sign = c.Query("sign")
	}

	if timestamp != "" && sign != "" {
		timestampInt64, _ := strconv.ParseInt(timestamp, 10, 64)
		nowTimestamp := time.Now().Unix()
		if (nowTimestamp-timestampInt64) > 300 || !sf.check(timestamp, path, appId, sign) {
			sf.SLogger().Error(fmt.Sprintf("403: timestamp: %s, path:%s, appId:%s, sign:%s", timestamp, path, appId, sign))
			c.JSON(403, map[string]any{"code": 403, "msg": "no permission"})
			c.Abort()
			return
		}
	} else {
		sf.SLogger().Error(fmt.Sprintf("403: timestamp: %s, sign:%s", timestamp, sign))
		c.JSON(403, map[string]any{"code": 403, "msg": "no permission"})
		c.Abort()
		return
	}
	// 处理请求
	c.Next()
}

func (sf *Sign) check(timestamp, path, appId, sign string) bool {
	str := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s%s", timestamp, path, appId)))
	fmt.Println(common.MD5(str), sign)
	return common.MD5(str) == sign
}
