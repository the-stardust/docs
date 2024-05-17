package app

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/controllers"
	"interview/router/request"
	"interview/services"
	"strconv"
)

type MorningRead struct {
	controllers.Controller
}

// 晨读信息
func (sf *MorningRead) InfoV2(c *gin.Context) {

	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")
	pageIndex := 1
	if pageIndexStr != "" {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
	}
	pageSize := 20
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}
	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))

	info, err, count := new(services.MorningReadService).GetMorningRead(c, offset, limit)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	for i, read := range info {
		if read.Cover == "https://xtj-question-bank.oss-cn-zhangjiakou.aliyuncs.com/morningRead/readDefault-cj66e3c32ur1df6tffs0.png" {
			info[i].Cover = "https://xtj-web.oss-cn-zhangjiakou.aliyuncs.com/static-file/xtj-admin/%E6%99%A8%E8%AF%BB%E6%A8%A1%E5%BC%8F%E5%B0%81%E9%9D%A2.png"
		}
	}
	list := map[string]any{
		"list":  info,
		"total": count,
	}

	sf.Success(list, c)
}

// 晨读信息
func (sf *MorningRead) Info(c *gin.Context) {

	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")
	pageIndex := 1
	if pageIndexStr != "" {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
	}
	pageSize := 100
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}
	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))
	info, err, _ := new(services.MorningReadService).GetMorningRead(c, offset, limit)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success(info, c)
}

// 晨读上报
func (sf *MorningRead) Report(c *gin.Context) {
	var err error
	var param = request.MorningReadReportRequest{}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	param.SourceType = sf.GetAppSourceType(c)
	err = new(services.MorningReadService).Report(uid, param)

	if err != nil {
		sf.SLogger().Error(err.Error())
		if err.Error() == "请完整作答" {
			sf.Error(common.CodeServerBusy, c, err.Error())
		} else {
			sf.Error(common.CodeServerBusy, c)
		}
		return
	}

	sf.Success("success", c)
}

// 晨读历史
func (sf *MorningRead) LogList(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")
	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")
	pageIndex := 1
	if pageIndexStr != "" {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
	}
	pageSize := 20
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}
	var filter = bson.M{"uid": uid}
	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))
	list, count, err := new(services.MorningReadService).LogList(filter, offset, limit)

	if err != nil {
		sf.Error(common.CodeServerBusy, c)
		return
	}

	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = count

	sf.Success(data, c)
}

// 晨读报告
func (sf *MorningRead) Result(c *gin.Context) {
	uid := c.GetHeader("X-XTJ-UID")

	morningReadId := c.Query("id")
	if morningReadId == "" {
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	info, err := new(services.MorningReadService).GetMorningReadReport(uid, morningReadId)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success(info, c)
}
