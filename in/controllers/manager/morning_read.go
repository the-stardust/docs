package manager

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"interview/common"
	"interview/controllers"
	"interview/params"
	"interview/router/request"
	"interview/services"
	"interview/services/word"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MorningRead struct {
	controllers.Controller
}

func (sf *MorningRead) MorningReadList(c *gin.Context) {
	id := c.Query("id")
	questionContent := c.Query("keyword")
	date := c.Query("date")
	state := c.Query("state")
	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")
	exam_category := c.Query("exam_category")
	exam_child_category := c.Query("exam_child_category")
	job_tag := c.Query("job_tag")
	province := c.Query("province")
	tag := c.Query("tag")

	var filter = bson.M{}
	if id != "" {
		filter["_id"] = sf.ObjectID(id)
	}
	if questionContent != "" {
		filterA := bson.A{bson.M{"question_content": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
			bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
			bson.M{"sub_name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(questionContent)}}},
		}
		if len(questionContent) == 24 {
			filterA = append(filterA, bson.M{"_id": sf.ObjectID(questionContent)})
		}
		filter["$or"] = filterA
	}
	if date != "" {
		filter["start_date"] = bson.M{"$lte": date}
		filter["end_date"] = bson.M{"$gte": date}
	}
	if state != "" {
		filter["state"], _ = strconv.Atoi(state)
	}
	if exam_category != "" {
		filter["exam_category"] = exam_category
	}
	if exam_child_category != "" {
		filter["exam_category"] = exam_category + "/" + exam_child_category
	}
	if job_tag != "" {
		filter["job_tag"] = job_tag
	}
	if province != "" {
		filter["province"] = province
	}
	if tag != "" {
		filter["tags"] = bson.M{"$in": []string{tag}}
	}
	pageIndex := 1
	if pageIndexStr != "" {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
	}
	pageSize := 20
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}

	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))
	list, count, err := new(services.MorningReadService).List(filter, offset, limit)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = count
	sf.Success(data, c)
}

// 保存
func (sf *MorningRead) MorningReadSave(c *gin.Context) {
	var err error
	var param = request.MorningReadSaveRequest{}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	if param.StartDate == "" {
		param.StartDate = time.Now().Format("2006-01-02")
	}

	for _, s := range param.QuestionAnswer {
		if s == "" || s == "undefined" {
			sf.Error(common.CodeServerBusy, c, "题目内容不能为空")
			return
		}
	}

	param.ManagerId = c.GetHeader("x-user-id")
	id, err := new(services.MorningReadService).Save(param)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success(id, c)
}

func (sf *MorningRead) MorningReadDel(c *gin.Context) {
	id := c.Query("id")

	err := new(services.MorningReadService).DelById(id)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}

	sf.Success("success", c)
}

func (sf *MorningRead) MorningReadItemLogList(c *gin.Context) {
	id := c.Query("id")
	uid := c.Query("uid")
	pageIndexStr := c.Query("page_index")
	pageSizeStr := c.Query("page_size")

	var filter = bson.M{}
	if id != "" {
		filter["morning_read_id"] = id
	}
	if uid != "" {
		filter["uid"] = uid
	}
	pageIndex := 1
	if pageIndexStr != "" {
		pageIndex, _ = strconv.Atoi(pageIndexStr)
	}
	pageSize := 20
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
	}

	offset, limit := sf.PageLimit(int64(pageIndex), int64(pageSize))
	list, count, err := new(services.MorningReadService).ItemLogList(filter, offset, limit)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = count
	sf.Success(data, c)
}

func (sf *MorningRead) MorningReadLogList(c *gin.Context) {
	var param params.MorningReadLogParam
	// 绑定参数
	if err := c.ShouldBind(&param); err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeInvalidParam, c)
		return
	}
	// 验证分页参数
	if param.PageIndex <= 0 {
		param.PageIndex = 1
	}
	if param.PageSize <= 0 {
		param.PageSize = 20
	}
	offset, limit := sf.PageLimit(param.PageIndex, param.PageSize)
	list, count, err := new(services.MorningReadService).MorningReadLogList(param, offset, limit)

	if err != nil {
		sf.SLogger().Error(err.Error())
		sf.Error(common.CodeServerBusy, c)
		return
	}
	data := make(map[string]interface{})
	data["list"] = list
	data["total"] = count
	sf.Success(data, c)
}

// 题目导入
func (sf *MorningRead) Import(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidUploadFile, c)
		return

	}
	defer file.Close()
	filename := header.Filename
	if path.Ext(filename) != ".docx" {
		sf.SLogger().Error("文件非word格式", filename)
		sf.Error(common.InvalidFileFormat, c)
		return
	}
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.InvalidFileFormat, c)
		return
	}
	// 初始化word解析器
	resolver := word.NewWordResolver()
	if len(buf.Bytes()) < 10 {
		sf.Error(common.CodeServerBusy, c, "无效的文件")
		return
	}
	//解析word内容
	err = resolver.Read(buf.Bytes())
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}
	data, err := resolver.GetWordData()
	if err != nil {
		sf.SLogger().Error(err)
		sf.Error(common.CodeServerBusy, c)
		return
	}

	// 定义一个结构体
	questions := make([]map[string]string, 0)
	for i := 0; i < len(data); i++ {
		question, index, errq := sf.getImportData(data, i)
		if len(errq) > 0 {
			errMsg := ""
			for _, eq := range errq {
				errMsg = fmt.Sprintf("  %s", eq)
			}
			sf.Error(common.CodeServerBusy, c, errMsg)
			return
		}
		if len(question) == 0 {
			break
		}
		questions = append(questions, question)
		i = index - 1
	}
	sf.SLogger().Info(fmt.Sprintf("解析了%d题", len(questions)))

	sf.Success(questions, c)
	return
}

func (sf *MorningRead) getImportData(data []string, index int) (map[string]string, int, []string) {
	var (
		getTitle bool
	)

	errorq := make([]string, 0)

	question := make(map[string]string)
	for i := index; i < len(data); i++ {
		content := data[i]
		index = i
		if strings.TrimSpace(content) == "" {
			continue
		}
		fmt.Printf(content)
		str, err := quickExerciseReg(content, `【素材】`)
		if err != nil {
			question["title"] += str
		} else {
			if getTitle {
				break
			}
			getTitle = true
			question["title"] = str
		}
	}
	return question, index, errorq
}

func quickExerciseReg(content, keyword string) (string, error) {
	reg, _ := regexp2.Compile(keyword, 0)
	regRes, err := reg.FindStringMatch(content)
	if err == nil && regRes != nil && regRes.String() != "" {
		return strings.Replace(content, regRes.String(), "", 1), nil
	}
	return content, errors.New(content)
}
