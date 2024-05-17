package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Upload struct {
	controllers.Controller
}

func (sf *Upload) RandomStr(c *gin.Context) {
	randStr := sf.randomStr()
	// randStr = randStr[16:] + randStr[:16] // 前端需要再次转过来
	sf.Success(randStr, c)
}

func (sf *Upload) randomStr() string {
	return common.MD5(strconv.Itoa(time.Now().Day()))
}

func (sf *Upload) checkSign(str, sign string) bool {
	return common.MD5(str) == sign
}

// 上传图片
func (sf *Upload) UploadFile(c *gin.Context) {
	var err error
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidUploadFile, c)
		return
	}
	defer file.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidFileFormat, c)
		return
	}
	objectName := c.PostForm("object_name")
	fileDownloadName := c.PostForm("file_download_name")
	bucketName := c.DefaultPostForm("bucket_name", "xtj-interview")
	url, err := new(services.Upload).UploadFile(objectName, fileDownloadName, buf.Bytes(), bucketName)
	if err != nil {

		sf.Error(common.CodeServerBusy, c)
		return
	}
	headerStr, _ := json.Marshal(c.Request.Header)
	sf.SLogger().Info(fmt.Sprintf("UploadFile file_name:%s file_size:%d header:%s url:%s", fileHeader.Filename, fileHeader.Size, headerStr, url))

	sf.Success(map[string]string{"url": url}, c)
}
func (sf *Upload) UploadFileBytes(c *gin.Context) {
	var err error
	var param struct {
		FileBytes        []byte `json:"file_bytes" binding:"required" `
		ObjectName       string ` json:"object_name"`
		FileDownloadName string `json:"file_download_name"`
		BucketName       string `json:"bucket_name"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.BucketName == "" {
		param.BucketName = "xtj-interview"
	}
	url, err := new(services.Upload).UploadFile(param.ObjectName, param.FileDownloadName, param.FileBytes, param.BucketName)
	if err != nil {

		sf.Error(common.CodeServerBusy, c)
		return
	}
	headerStr, _ := json.Marshal(c.Request.Header)
	sf.SLogger().Info(fmt.Sprintf("UploadFileBytes ObjectName:%s FileDownloadName:%d header:%s url:%s", param.ObjectName, param.FileDownloadName, headerStr, url))
	sf.Success(map[string]string{"url": url}, c)
}

// 学员上传试题文件
func (sf *Upload) UploadFile2(c *gin.Context) {
	var err error
	uid := c.GetHeader("X-XTJ-UID")
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidUploadFile, c)
		return
	}
	defer file.Close()
	objectName := c.DefaultPostForm("object_name", "student/files/")
	objectName += fileHeader.Filename
	fileDownloadName := c.DefaultPostForm("file_download_name", fileHeader.Filename)
	bucketName := c.DefaultPostForm("bucket_name", "xtj-interview")
	examCategory := c.DefaultPostForm("exam_category", "")
	examChildCategory := c.DefaultPostForm("exam_child_category", "")

	// sign验证
	sign := c.PostForm("sign")
	if !sf.checkSign(fmt.Sprintf("%s%s%s%s", uid, objectName, uid, sf.randomStr()), sign) {
		sf.SLogger().Error(fmt.Sprintf("uid:%s", uid))
		sf.Error(common.NoPermission, c)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidFileFormat, c)
		return
	}
	url, err := new(services.Upload).UploadFile(objectName, fileDownloadName, buf.Bytes(), bucketName)
	if err != nil {
		sf.Error(common.CodeServerBusy, c)
		return
	}
	headerStr, _ := json.Marshal(c.Request.Header)
	sf.SLogger().Info(fmt.Sprintf("UploadFile file_name:%s file_size:%d header:%s url:%s", fileHeader.Filename, fileHeader.Size, headerStr, url))

	// 保存用户的文件上传记录
	var userUploadFile models.UserUploadFile
	userUploadFile.ExamCategory = examCategory
	userUploadFile.ExamChildCategory = examChildCategory
	userUploadFile.UserID = uid
	userUploadFile.FileUrl = url
	userUploadFile.FileName = fileHeader.Filename
	userUploadFile.CheckStatus = 0
	_, err = sf.DB().Collection("user_upload_file").Create(&userUploadFile)
	if err != nil {
		sf.Error(common.CodeServerBusy, c)
		return
	}
	sf.Success(map[string]string{"url": url}, c)
}
