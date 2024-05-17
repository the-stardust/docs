package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/controllers"
	"interview/models"
	"interview/services"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ImaController struct {
	controllers.Controller
}

type ImaReqParams struct {
	UserName          string                 `json:"user_name"`
	Mobile            string                 `json:"mobile"`
	InterviewMockID   string                 `json:"interview_mock_id"`
	OneByOneTimeType  int                    `json:"onebyone_time_type" bson:"onebyone_time_type"`
	SimulateTimeType  int                    `json:"simulate_time_type" bson:"simulate_time_type"`
	QuestionnaireInfo []models.Questionnaire `json:"questionnaire_info,omitempty"`
}

type ExportParams struct {
	ID         string `json:"id"`
	GroupType  int    `json:"group_type"`
	GroupParam int    `json:"group_param"`
}

type excelItem struct {
	name            string
	userNameColor   []colorItem
	className2Color map[string]string
}

type colorItem struct {
	color string
	value string
}

func (i *ImaController) ApplyUserList(c *gin.Context) {
	id, ok := c.GetQuery("id")
	if !ok {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	f := bson.M{"interview_mock_id": id}
	list, err := models.ImaModel.GetApplyListByFilter(f)
	if err != nil {
		i.Error(common.CodeServerBusy, c)
		return
	}
	type resp struct {
		UserID           string `json:"user_id"`
		UserName         string `json:"user_name"`
		ClassName        string `json:"class_name"`
		Mobile           string `json:"mobile"`
		OneByOneTimeType int    `json:"onebyone_time_type" bson:"onebyone_time_type"`
		SimulateTimeType int    `json:"simulate_time_type" bson:"simulate_time_type"`
	}
	res := make([]resp, 0, len(list))
	uids := make([]string, 0, len(list))
	repeatNameCount := make(map[string]int)
	for _, v := range list {
		uids = append(uids, v.UserID)
		tmp := resp{
			UserID:           v.UserID,
			UserName:         v.UserName,
			ClassName:        v.ClassName,
			Mobile:           v.Mobile,
			OneByOneTimeType: v.OneByOneTimeType,
			SimulateTimeType: v.SimulateTimeType,
		}
		if _, ok := repeatNameCount[v.UserName]; ok {
			repeatNameCount[v.UserName]++
		} else {
			repeatNameCount[v.UserName] = 1
		}
		res = append(res, tmp)
	}
	// 查询 class
	classMap, _ := models.InterviewClassModel.GetUsersClassMap(uids)
	for index := range res {
		tmp := res[index]
		if v, ok := classMap[tmp.UserID]; ok {
			tmp.ClassName = v.Name
		}
		// 重名处理
		if repeatNameCount[tmp.UserName] > 1 && len(tmp.Mobile) > 4 {
			tmp.UserName = tmp.UserName + tmp.Mobile[len(tmp.Mobile)-4:]
		}
		res[index] = tmp
	}
	i.Success(res, c)
	return
}

// IsApply 获取用户某个模考的报名状态
func (i *ImaController) IsApply(c *gin.Context) {
	var err error
	var params ImaReqParams
	err = c.ShouldBindJSON(&params)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c)
		return
	}
	userID := c.GetHeader("X-XTJ-UID")
	if userID == "" {
		i.Error(common.CodeInvalidParam, c, "un login")
		return
	}
	applyInfo, err := models.ImaModel.GetApplyInfoByUser(userID, params.InterviewMockID)
	if err != nil {
		i.Error(common.CodeInvalidParam, c, "un login")
		return
	}
	flag := false
	if applyInfo != nil && applyInfo.Status == 1 {
		flag = true
	}
	i.Success(flag, c)

}

// ApplyMock 报名模考 除去考试管理员 任何人都能报名
func (i *ImaController) ApplyMock(c *gin.Context) {
	var err error
	var param ImaReqParams
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c)
		return
	}
	if param.InterviewMockID == "" {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	if param.OneByOneTimeType == 0 && param.SimulateTimeType == 0 {
		i.Error(common.CodeInvalidParam, c, "请选择参数方式和时间")
		return
	}
	userID := c.GetHeader("X-XTJ-UID")
	if userID == "" {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	// 获取用户是否是考试管理员
	userControlCurrIDs := models.CurriculaModel.GetUserControlCurrID(userID)
	if len(userControlCurrIDs) > 0 {
		i.Error(common.PermissionDenied, c, "考试管理员不允许报名")
		return
	}
	var isRepeat bool
	var class *models.InterviewClass
	var mockInfo *models.InterviewMock
	var mockInfoErr error
	var wg sync.WaitGroup
	// 获取已报名信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		isRepeat = models.ImaModel.CheckRepeatApply(param.InterviewMockID, userID, param.UserName, param.Mobile)
	}()
	// 获取模考信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		mockInfo, mockInfoErr = models.InterviewMockModel.GetInfo(i.ObjectID(param.InterviewMockID))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		class, _ = models.InterviewClassModel.GetUserClass(userID)
	}()
	wg.Wait()
	if mockInfoErr != nil {
		i.Error(common.CodeInvalidParam, c, mockInfoErr.Error())
		return
	}
	location, _ := time.LoadLocation("Asia/Shanghai")
	// 判断报名是否结束
	end, err := time.ParseInLocation("2006-01-02 15:04", mockInfo.ExamEndTime, location)
	if err != nil {
		i.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	if time.Now().After(end) {
		i.Error(common.CodeInvalidParam, c, "报名已结束")
		return
	}
	if isRepeat {
		i.Error(common.CodeInvalidParam, c, "重复报名")
		return
	}

	info := models.InterviewMockApply{
		UserID:            userID,
		UserName:          param.UserName,
		Mobile:            param.Mobile,
		InterviewMockID:   param.InterviewMockID,
		OneByOneTimeType:  param.OneByOneTimeType,
		SimulateTimeType:  param.SimulateTimeType,
		QuestionnaireInfo: param.QuestionnaireInfo,
	}
	if class != nil {
		info.ClassID = class.Id.Hex()
		info.ClassName = class.Name
	}
	err = models.ImaModel.Create(info)
	if err != nil {
		i.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	i.Success(nil, c)
}

// ApplyCancel 取消报名模考
func (i *ImaController) ApplyCancel(c *gin.Context) {
	var err error
	var param struct {
		ID string `json:"id"`
	}
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	userID := c.GetHeader("X-XTJ-UID")
	err = models.ImaModel.ApplyCancel(userID, param.ID)
	if err != nil {
		i.SLogger().Error(err.Error())
		i.Error(common.CodeInvalidParam, c, err.Error())
		return
	}
	i.Success(true, c)
}

// Export 导出模考表格
func (i *ImaController) Export(c *gin.Context) {
	var err error
	var param ExportParams
	err = c.ShouldBindJSON(&param)
	if err != nil {
		i.Error(common.CodeInvalidParam, c)
		return
	}
	if param.GroupParam <= 0 {
		i.Error(common.CodeInvalidParam, c, "分组参数需要大于0")
		return
	}
	userID := c.GetHeader("X-XTJ-UID")
	// 模考info
	mockInfo, mockInfoErr := models.InterviewMockModel.GetInfo(i.ObjectID(param.ID))
	if mockInfoErr != nil {
		i.Error(common.CodeInvalidParam, c, mockInfoErr.Error())
		return
	}
	if mockInfo == nil {
		i.Error(common.CodeInvalidParam, c, "模考不存在")
		return
	}
	if !models.CurriculaModel.CheckCurriculaAdmin(userID, mockInfo.CurriculaID) {
		i.Error(common.PermissionDenied, c, "不是考试管理员不可导出")
		return
	}
	url, err := export(mockInfo, param.GroupType, param.GroupParam)
	if err != nil {
		i.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	i.Success(url, c)
}

func getUserData(mockInfo *models.InterviewMock) map[string][]models.InterviewMockApply {
	// 上午和下午的模块是两场 one_by_one 和 simulate 也要分开
	users := make(map[string][]models.InterviewMockApply)
	var wg sync.WaitGroup
	var oneByoneAM, oneByonePM, simulateAM, simulatePM []models.InterviewMockApply
	if mockInfo.OneByOneTimeType != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 先查 one_by_one
			if mockInfo.OneByOneTimeType == 3 {
				oneByoneAM, _ = models.ImaModel.GetApplyListByFilter(bson.M{"interview_mock_id": mockInfo.Id.Hex(), "onebyone_time_type": bson.M{"$in": []int{1, 3}}})
				oneByonePM, _ = models.ImaModel.GetApplyListByFilter(bson.M{"interview_mock_id": mockInfo.Id.Hex(), "onebyone_time_type": bson.M{"$in": []int{2, 3}}})
			} else {
				f := bson.M{"interview_mock_id": mockInfo.Id.Hex(), "onebyone_time_type": mockInfo.OneByOneTimeType}
				if mockInfo.OneByOneTimeType == 2 {
					oneByonePM, _ = models.ImaModel.GetApplyListByFilter(f)
				} else {
					oneByoneAM, _ = models.ImaModel.GetApplyListByFilter(f)
				}
			}
		}()
	}
	if mockInfo.SimulateTimeType != 0 {
		// 再查询全真模拟
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 先查 one_by_one
			if mockInfo.SimulateTimeType == 3 {
				simulateAM, _ = models.ImaModel.GetApplyListByFilter(bson.M{"interview_mock_id": mockInfo.Id.Hex(), "simulate_time_type": bson.M{"$in": []int{1, 3}}})
				simulatePM, _ = models.ImaModel.GetApplyListByFilter(bson.M{"interview_mock_id": mockInfo.Id.Hex(), "simulate_time_type": bson.M{"$in": []int{2, 3}}})
			} else {
				f := bson.M{"interview_mock_id": mockInfo.Id.Hex(), "simulate_time_type": mockInfo.SimulateTimeType}
				if mockInfo.SimulateTimeType == 2 {
					simulatePM, _ = models.ImaModel.GetApplyListByFilter(f)
				} else {
					simulateAM, _ = models.ImaModel.GetApplyListByFilter(f)
				}
			}
		}()
	}
	wg.Wait()
	users["一对一模拟上午分组"] = oneByoneAM
	users["一对一模拟下午分组"] = oneByonePM
	users["全真模拟上午分组"] = simulateAM
	users["全真模拟下午分组"] = simulatePM
	return users
}

func export(mockInfo *models.InterviewMock, groupType, groupParam int) (string, error) {
	users := getUserData(mockInfo)
	f := excelize.NewFile()
	// one_by_one 和 simulate 分开 sheet 导
	index := 1
	for title, user := range users {
		if len(user) == 0 {
			continue
		}
		sheetData := makeData(user, groupType, groupParam)
		originSheet := "Sheet" + strconv.Itoa(index)
		err := f.SetSheetName(originSheet, title)
		if err != nil {
			continue
		}
		generateExcel(f, title, sheetData)
		index++
	}
	// 保存Excel文件
	fileName := "模考分组" + time.Now().Format("20060102150405") + ".xlsx"
	err := f.SaveAs(fileName)
	if err != nil {
		return "", err
	}
	defer os.Remove(fileName)
	// 上传阿里云oss
	fileBytes, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	fileUrl, err := new(services.Upload).UploadFile(fmt.Sprintf("/interview-mock/%s", fileName), "", fileBytes)
	if err != nil {
		return "", err
	}
	return fileUrl, nil
}

func generateExcel(f *excelize.File, sheetName string, data []excelItem) {
	res := make([][]string, len(data))
	for i := range data {
		res[i] = make([]string, len(res[i]))
	}
	if len(data) == 0 {
		return
	}
	index, _ := f.GetSheetIndex(sheetName)
	if index < 0 {
		_, _ = f.NewSheet(sheetName)
	}
	// 设置表头格式
	styleCenter := centerStyle(f)
	_ = f.SetRowHeight(sheetName, 1, 40)
	_ = f.SetRowHeight(sheetName, 2, 30)
	_ = f.SetColWidth(sheetName, "A", "K", 30)
	_ = f.SetCellValue(sheetName, "A1", "分组安排")
	_ = f.SetCellStyle(sheetName, "A1", "A1", styleCenter)
	_ = f.MergeCell(sheetName, "A1", fmt.Sprintf("%s1", string('A'+int32(len(data)-1))))

	classNameStart := len(data[0].userNameColor) + 5
	// 按单元格写数据与背景颜色
	for i := 'A'; i < 'A'+int32(len(res)); i++ {
		dataIndex := i - 'A'
		_ = f.SetCellValue(sheetName, fmt.Sprintf("%s2", string(i)), fmt.Sprintf("第%d组", dataIndex+1))
		_ = f.SetCellStyle(sheetName, fmt.Sprintf("%s2", string(i)), fmt.Sprintf("%s2", string(i)), styleCenter)
		for j := 3; j <= len(data[dataIndex].userNameColor)+2; j++ {
			_ = f.SetRowHeight(sheetName, j, 20)
			cell := fmt.Sprintf("%s%d", string(i), j)
			val := data[dataIndex].userNameColor[j-3]

			style := contentStyle(f, val.color)
			_ = f.SetCellValue(sheetName, cell, val.value)
			_ = f.SetCellStyle(sheetName, cell, cell, style)
		}

		start := 0
		for value, color := range data[dataIndex].className2Color {
			cell := fmt.Sprintf("%s%d", string(i), classNameStart+start)
			style := contentStyle(f, color)
			_ = f.SetRowHeight(sheetName, classNameStart+start, 20)
			_ = f.SetCellValue(sheetName, cell, value)
			_ = f.SetCellStyle(sheetName, cell, cell, style)
			start++
		}

	}
}

func makeData(users []models.InterviewMockApply, groupType, groupParam int) []excelItem {
	if len(users) == 0 {
		return nil
	}
	bucket := groupParam
	max := int(math.Ceil(float64(len(users)) / float64(groupParam)))
	// group_type = 1时 会固定生成group_param个组， group_type = 2 时，会按每group_param个人分成一组
	// 先确定桶的数量和每个桶最多多少人
	if groupType == 2 {
		bucket = int(math.Ceil(float64(len(users)) / float64(groupParam)))
		max = groupParam
	}
	// 获取班级颜色
	colorMap := map[string]string{}
	userIDs := make([]string, 0, len(users))
	repeatNameCount := make(map[string]int)
	for _, user := range users {
		// 重名判断
		if _, ok := repeatNameCount[user.UserName]; ok {
			repeatNameCount[user.UserName]++
		} else {
			repeatNameCount[user.UserName] = 1
		}
		userIDs = append(userIDs, user.UserID)
	}
	classMap, _ := models.InterviewClassModel.GetUsersClassMap(userIDs)
	for i := range users {
		tmp := users[i]
		if v, ok := classMap[tmp.UserID]; ok {
			tmp.ClassName = v.Name
		}
		users[i] = tmp
		colorMap[tmp.ClassName] = randColor()
	}
	// 几个班级就有几种颜色
	l := bucket
	if l > len(users) {
		l = len(users)
	}
	result := make([]excelItem, bucket)
	for i := range result {
		result[i].className2Color = make(map[string]string)
		result[i].userNameColor = make([]colorItem, 0, max)
	}
	bucketIndex := 0
	for _, user := range users {
		tmp := result[bucketIndex]
		color := colorMap[user.ClassName]
		// 学生名字填充 重名要加手机号
		userName := user.UserName
		if repeatNameCount[user.UserName] > 1 && len(user.Mobile) > 4 {
			userName += user.Mobile[len(user.Mobile)-4:]
		}
		tmp.userNameColor = append(tmp.userNameColor, colorItem{
			color: color,
			value: userName,
		})
		// 班级名称填充
		tmp.className2Color[user.ClassName] = color
		result[bucketIndex] = tmp
		// 跳桶
		if len(tmp.userNameColor) == max {
			bucketIndex++
		}
	}
	return result
}

func randColor() string {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())
	// 生成一个随机的RGB颜色值
	arr := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}
	var res strings.Builder
	for i := 0; i < 6; i++ {
		res.WriteString(arr[rand.Intn(len(arr))])
	}
	return res.String()
}

func centerStyle(f *excelize.File) int {
	styleCenter, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   18,
			Family: "宋体",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	return styleCenter
}

func contentStyle(f *excelize.File, color string) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "宋体",
		},
		Fill: excelize.Fill{
			Type:  "gradient",
			Color: []string{color, color},
		},
	})
	return style
}
