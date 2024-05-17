package common

type ResCode int64

const (
	CodeSuccess      ResCode = 0
	CodeInvalidParam ResCode = 1000 + iota
	CodeServerBusy
	PermissionDenied
	InvalidId
	ForbidDel
	ForbidUpdate
	InvalidUploadFile
	InvalidFileFormat
	InvalidFormat
	Closed
	Expired
	DuplicateQuestion
	NoPermission
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:       "success",
	CodeInvalidParam:  "请求参数错误",
	CodeServerBusy:    "服务器繁忙",
	PermissionDenied:  "权限不足",
	InvalidId:         "无效得ID",
	ForbidDel:         "禁止删除",
	ForbidUpdate:      "禁止编辑",
	InvalidUploadFile: "缺少上传文件",
	InvalidFileFormat: "无效的文件格式",
	InvalidFormat:     "格式异常",
	Closed:            "已关闭",
	Expired:           "已过期",
	DuplicateQuestion: "重复试题",
	NoPermission:      "没有权限！",
}

func (c ResCode) GetMsg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
