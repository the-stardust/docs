package services

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common"
	"interview/models"
	"sort"
	"strings"
	"time"
)

/**
* @Author XuPEngHao
* @Date 2023/12/7 18:47
 */

type categorySrv struct {
	ServicesBase
}

func NewCategorySrv() *categorySrv {
	return new(categorySrv)
}

func (c *categorySrv) GetQuestionCategoryItem(examCategory, examChildCategory string) ([]models.QuestionCategoryItem, error) {
	var qc models.QuestionCategory
	filter := bson.M{"exam_category": examCategory}
	if examChildCategory != "" {
		filter["exam_child_category"] = examChildCategory
	}
	err := c.DB().Collection("question_category").Where(filter).Take(&qc)
	return qc.Categorys, err
}

func (c *categorySrv) BuildCateCountMap(questions []models.GQuestion) map[string]map[string]models.QuestionSimpleItem {
	res := make(map[string]map[string]models.QuestionSimpleItem)

	for _, q := range questions {
		allCate := ""
		if len(q.QuestionCategory) == 0 {
			allCate = "全部"
		} else {
			allCate = strings.Join(q.QuestionCategory, "_")
		}
		tmp := models.QuestionSimpleItem{
			QuestionID: q.Id.Hex(),
			CreateTime: q.CreatedTime,
			Year:       q.Year,
			Month:      q.Month,
			Day:        q.Day,
		}
		if _, ok := res[allCate]; ok {
			res[allCate][tmp.QuestionID] = tmp
		} else {
			res[allCate] = make(map[string]models.QuestionSimpleItem)
			res[allCate][tmp.QuestionID] = tmp
		}
	}
	return res
}

func AddAnswerQuestionIDs(root *models.KeypointStatisticsResp, answerMap map[string][]string) {
	if root == nil {
		return
	}
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i, c := range root.Child {
		if v, ok := answerMap[c.AllCate]; ok {
			(*root).Child[i].AnswerQuestionID = v
			(*root).Child[i].AnswerCount = len(v)
		}
		if (*root).Child[i].Child == nil || len((*root).Child[i].Child) == 0 {
			continue
		}
		AddAnswerQuestionIDs(&(*root).Child[i], answerMap)
	}
}
func AddQuestionIDs(root *models.KeypointStatisticsResp, allQuestionCateMap map[string]map[string]models.QuestionSimpleItem) {
	if root == nil {
		return
	}
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i, c := range root.Child {
		if v, ok := allQuestionCateMap[c.AllCate]; ok {
			(*root).Child[i].AllQuestionIDMap = v
		}
		if (*root).Child[i].Child == nil || len((*root).Child[i].Child) == 0 {
			continue
		}
		AddQuestionIDs(&(*root).Child[i], allQuestionCateMap)
	}
}

func LastViewID(root *models.KeypointStatisticsResp, record map[string]string) {
	if root == nil {
		return
	}
	if v, ok := record[root.AllCate]; ok {
		// get qid
		var info models.UserRecordValue
		err := json.Unmarshal([]byte(v), &info)
		if err == nil {
			root.LastViewQuestionID = info.LastViewQuestionID
		}
	}
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i := range root.Child {
		LastViewID(&root.Child[i], record)
	}
}

func SortKeypoint(root *models.KeypointStatisticsResp) {
	if root == nil {
		return
	}
	keypointSort := []string{
		"全部",
		"社会现象",
		"态度观点",
		"组织管理",
		"应急应变",
		"人际关系",
		"情景模拟",
		"开放论述",
		"演讲/串词",
		"漫画",
		"自我认知",
		"复合化题型",
		"专业题",
	}
	newChild := make([]models.KeypointStatisticsResp, 0, len(root.Child))
	findFunc := func(target string, child []models.KeypointStatisticsResp) models.KeypointStatisticsResp {
		for _, v := range child {
			if target == v.Title {
				return v
			}
		}
		return models.KeypointStatisticsResp{}
	}
	for _, k := range keypointSort {
		res := findFunc(k, root.Child)
		if res.Title != k {
			continue
		}
		newChild = append(newChild, res)
	}
	root.Child = newChild
}

func HasNew(root *models.KeypointStatisticsResp, record map[string]string) {
	if root == nil {
		return
	}
	if v, ok := record[root.AllCate]; ok && len(root.AllQuestionID) > 0 {
		// check index
		var info models.UserRecordValue
		err := json.Unmarshal([]byte(v), &info)
		if err == nil {
			if info.LastViewQuestionID == "" {
				root.HasNew = true
			} else {
				oldIndex := info.LastViewQuestionIDIndex
				newIndex := FindIndex(info.LastViewQuestionID, root.AllQuestionID)

				if newIndex > oldIndex {
					root.HasNew = true
				}
			}
		}
	}
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i := range root.Child {
		HasNew(&root.Child[i], record)
	}
}

func FindIndex(target string, arr []string) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return 0
}

func SortQuestion(root *models.KeypointStatisticsResp) {
	if root == nil {
		return
	}
	// pai xu
	tmp := make([]models.QuestionSimpleItem, 0, len(root.AllQuestionIDMap))
	for _, q := range root.AllQuestionIDMap {
		tmp = append(tmp, q)
	}
	sort.Slice(tmp, func(i, j int) bool {
		// 先按年月日排
		if tmp[i].Year != tmp[j].Year {
			return tmp[i].Year > tmp[j].Year
		}
		if tmp[i].Month != tmp[j].Month {
			return tmp[i].Month > tmp[j].Month
		}
		if tmp[i].Day != tmp[j].Day {
			return tmp[i].Day > tmp[j].Day
		}
		createI, _ := time.Parse("2006-01-02 15:04:05", tmp[i].CreateTime)
		createJ, _ := time.Parse("2006-01-02 15:04:05", tmp[j].CreateTime)
		return createI.After(createJ)
	})
	allIDs := make([]string, 0, len(tmp))
	for _, q := range tmp {
		allIDs = append(allIDs, q.QuestionID)
	}
	root.AllQuestionID = allIDs
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	for i := range root.Child {
		SortQuestion(&root.Child[i])
	}
}
func AllTabAnswerCount(root *models.KeypointStatisticsResp) {
	if root.Child == nil || len(root.Child) == 0 {
		return
	}
	if root.Child[0].Title != "全部" {
		return
	}
	allAnsIDs := make([]string, 0)
	for i := 1; i < len(root.Child); i++ {
		allAnsIDs = append(allAnsIDs, root.Child[i].AnswerQuestionID...)
	}
	if len(allAnsIDs) == 0 {
		return
	}
	root.Child[0].AnswerQuestionID = allAnsIDs
	root.Child[0].AnswerCount = len(allAnsIDs)
}

func AddAllTab(root *models.KeypointStatisticsResp, allQuestionCateMap map[string]map[string]models.QuestionSimpleItem) {
	allIDs := make(map[string]models.QuestionSimpleItem)
	if v, ok := allQuestionCateMap["全部"]; ok {
		allIDs = v
	}
	for _, q := range root.Child {
		mergeQuestionIDMap(allIDs, q.AllQuestionIDMap)
	}
	if len(allIDs) == 0 {
		return
	}
	newChild := []models.KeypointStatisticsResp{
		{
			Title:            "全部",
			AllCate:          "全部",
			AllQuestionIDMap: allIDs,
			AllQuestionCount: len(allIDs),
		},
	}
	newChild = append(newChild, root.Child...)
	root.Child = newChild
}

func DfsQuestionCount(root *models.KeypointStatisticsResp) {
	if root == nil {
		return
	}
	if root.Child == nil || len(root.Child) == 0 {
		root.AllQuestionCount = len(root.AllQuestionIDMap)
		return
	}
	allIds := make(map[string]models.QuestionSimpleItem)
	for i := range root.Child {
		DfsQuestionCount(&root.Child[i])
		mergeQuestionIDMap(allIds, root.Child[i].AllQuestionIDMap)
	}
	if len(root.AllQuestionIDMap) == 0 || root.AllQuestionIDMap == nil {
		root.AllQuestionIDMap = make(map[string]models.QuestionSimpleItem)
	}
	mergeQuestionIDMap(root.AllQuestionIDMap, allIds)
	root.AllQuestionCount = len(root.AllQuestionIDMap)
}

func DfsAnswerCount(root *models.KeypointStatisticsResp) {
	if root == nil {
		return
	}
	if root.Child == nil || len(root.Child) == 0 {
		root.AnswerCount = len(root.AnswerQuestionID)
		return
	}
	ansIDs := make([]string, 0, root.AnswerCount)
	for i := range root.Child {
		DfsAnswerCount(&root.Child[i])
		ansIDs = common.AppendSet(ansIDs, root.Child[i].AnswerQuestionID)
	}
	root.AnswerQuestionID = common.AppendSet(root.AnswerQuestionID, ansIDs)
	root.AnswerCount = len(root.AnswerQuestionID)
}
func mergeQuestionIDMap(allIds map[string]models.QuestionSimpleItem, questionIDs map[string]models.QuestionSimpleItem) {
	if len(questionIDs) == 0 {
		return
	}
	for i := range questionIDs {
		allIds[questionIDs[i].QuestionID] = questionIDs[i]
	}
}

func BuildTree(questions []models.GQuestion) []models.KeypointStatisticsResp {
	root := models.KeypointStatisticsResp{
		Child: make([]models.KeypointStatisticsResp, 0),
	}
	for _, q := range questions {
		build(&root, q.QuestionCategory, 0)
	}
	return root.Child
}

func build(root *models.KeypointStatisticsResp, c []string, cIndex int) {
	if root == nil {
		root = &models.KeypointStatisticsResp{}
	}
	if cIndex >= len(c) {
		return
	}
	childIndex := root.FindChild(c[cIndex])
	if childIndex != -1 {
		cIndex++
		build(&root.Child[childIndex], c, cIndex)
		return
	}
	tmp := models.KeypointStatisticsResp{
		Title:   c[cIndex],
		AllCate: strings.Join(c[:cIndex+1], "_"),
		Child:   nil,
	}
	(*root).Child = append((*root).Child, tmp)
	cIndex++
	i := len(root.Child) - 1
	build(&root.Child[i], c, cIndex)
}
