package app

import (
	"github.com/gin-gonic/gin"
	"interview/common"
	"interview/controllers"
	"interview/params"
	"interview/services"
)

type CurriculaControl struct {
	controllers.Controller
	service *services.CurriculaSrv
}

func GetInitControl() *CurriculaControl {
	return &CurriculaControl{
		controllers.Controller{},
		services.NewCurriculaSrv(),
	}
}
func (sf CurriculaControl) CurriculaIsShow(c *gin.Context) {
	userId := c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(userId, "")
	res := map[string]bool{
		"is_show": flag,
	}
	sf.Success(res, c)
}

// CurriculaList 课程列表
func (sf *CurriculaControl) CurriculaList(c *gin.Context) {
	var param params.CurriculaListRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.UserId = c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(param.UserId, "")
	if !flag {
		sf.SLogger().Error("权限不足")
		sf.Error(common.PermissionDenied, c)
		return
	}
	data, err := sf.service.QueryCurriculaListService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

// CurriculaSelectList 课程列表不需要权限
func (sf *CurriculaControl) CurriculaSelectList(c *gin.Context) {
	var param params.CurriculaListRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	data, err := sf.service.QueryCurriculaSelectListService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

func (sf CurriculaControl) CurriculaSave(c *gin.Context) {
	var param params.CurriculaRequestParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.CreateUserId = c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(param.CreateUserId, param.Id)
	if !flag {
		sf.SLogger().Error("权限不足")
		sf.Error(common.PermissionDenied, c)
		return
	}
	err := sf.service.CreateCurriculaService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

func (sf CurriculaControl) CurriculaInviteCode(c *gin.Context) {
	var param params.CurriculaInviteCodeRequestParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.UserId = c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(param.UserId, param.Id)
	if !flag {
		sf.SLogger().Error("权限不足")
		sf.Error(common.PermissionDenied, c)
		return
	}
	data, err := sf.service.CurriculaInviteCodeService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

func (sf CurriculaControl) CheckCurriculaAdmin(c *gin.Context) {
	var param params.CurriculaInviteCodeRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.UserId = c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(param.UserId, param.Id)
	res := map[string]interface{}{
		"is_admin": flag,
		"id":       param.Id,
	}
	sf.Success(res, c)
}

func (sf CurriculaControl) CurriculaInviteUse(c *gin.Context) {
	var param params.CurriculaInviteCodeUseRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.UserId = c.GetHeader("X-XTJ-UID")
	data, err := sf.service.CurriculaInviteUseService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

func (sf CurriculaControl) CurriculaAdminUserDelete(c *gin.Context) {
	var param params.CurriculaAdminUserIdRequestParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.ActionUserId = c.GetHeader("X-XTJ-UID")
	err := sf.service.CurriculaAdminUserDeleteService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
}

func (sf CurriculaControl) CurriculaAdminUserList(c *gin.Context) {
	var param params.CurriculaAdminListRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	param.ActionUserId = c.GetHeader("X-XTJ-UID")
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(param.ActionUserId, param.Id)
	if !flag {
		sf.SLogger().Error("权限不足")
		sf.Error(common.PermissionDenied, c)
		return
	}
	param.AppCode = c.GetString("APP-CODE")
	data, err := sf.service.CurriculaAdminUserListService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

// AdminClassList 面试班列表
func (sf CurriculaControl) AdminClassList(c *gin.Context) {
	var param params.AdminClassListRequestParam
	if err := c.ShouldBindQuery(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// uid := c.GetHeader("X-XTJ-UID")
	// 检测权限
	//flag := sf.service.CheckCurriculaAdminRedis(uid, param.CurriculaId)
	//if !flag {
	//	sf.SLogger().Error("权限不足")
	//	sf.Error(common.PermissionDenied, c, "权限不足")
	//	return
	//}
	data, err := sf.service.AdminClassListService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(data, c)
}

// AdminSaveInterviewQuestion 保存试题
func (sf CurriculaControl) AdminSaveInterviewQuestion(c *gin.Context) {
	var param params.AdminSaveInterviewQuestionRequestParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	uid := c.GetHeader("X-XTJ-UID")
	param.UserId = uid
	// 检测权限
	flag := sf.service.CheckCurriculaAdminRedis(uid, param.CurriculaId)
	if !flag {
		sf.SLogger().Error("权限不足")
		sf.Error(common.PermissionDenied, c, "权限不足")
		return
	}
	err := sf.service.AdminSaveInterviewQuestionService(param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(nil, c)
	return
}
