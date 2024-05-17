package jobs

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/models"
	"interview/services"
	"regexp"
	"sync"
	"time"
)

func init() {
	keypointCountJob := new(KeypointCountJob)
	RegisterWorkConnectJob(keypointCountJob, 90, 5)
}

type KeypointCountJob struct {
	JobBase
}

var jobTag = []string{
	"税务局",
	"公安联考",
	"海关",
	"铁路公安",
	"外交部",
	"银保监会",
	"海事",
	"统计局调查总队",
	"部位党群及参公单位",
}

var areas = []string{
	"国家",
	"安徽",
	"北京",
	"重庆",
	"福建",
	"甘肃",
	"贵州",
	"广西",
	"广东",
	"河南",
	"湖南",
	"黑龙江",
	"湖北",
	"河北",
	"海南",
	"江苏",
	"江西",
	"吉林",
	"辽宁",
	"内蒙古",
	"宁夏",
	"青海",
	"山东",
	"四川",
	"陕西",
	"山西",
	"上海",
	"天津",
	"新建",
	"西藏",
	"云南",
	"浙江",
}

var provenceExam = []string{
	"事业单位",
	"选调生",
}

func (k KeypointCountJob) GetJobs() ([]interface{}, error) {
	var examCategorys = make([]models.ExamCategory, 0)
	err := k.DB().Collection("exam_category").Sort("-sort_number").Find(&examCategorys)
	if err != nil {
		return []interface{}{}, err
	}
	var items []interface{}
	for _, v := range examCategorys {
		items = append(items, v)
	}
	return items, nil
}

func (k KeypointCountJob) Do(item interface{}) error {
	examCategory := item.(models.ExamCategory)
	// 公务员处理 provence  和 job_tag
	if examCategory.Title == "公务员" {
		k.dealJobTag(examCategory)
		k.dealProvence(examCategory)
		k.dealCommon(examCategory, "", "")
	} else if common.InArrCommon(examCategory.Title, provenceExam) {
		k.dealProvence(examCategory)
		k.dealCommon(examCategory, "", "")
	} else {
		k.dealCommon(examCategory, "", "")
	}
	return nil
}

func (k KeypointCountJob) dealCommon(examCategory models.ExamCategory, jobTag, provence string) {
	if len(examCategory.ChildCategory) > 0 {
		for _, v := range examCategory.ChildCategory {
			k.deal(examCategory.Title, v.Title, jobTag, provence)
		}
	} else {
		k.deal(examCategory.Title, "", jobTag, provence)
	}
}

// deal
func (k KeypointCountJob) build(filter bson.M) []models.KeypointStatisticsResp {
	var questions []models.GQuestion
	err := k.DB().Collection("g_interview_questions").Where(filter).Find(&questions)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// 获取 question_category 对应的 question_ids
	allQuestionCateMap := services.NewCategorySrv().BuildCateCountMap(questions)
	//  把所有 question 的分类 build 成多叉树,形成多级分类
	res := services.BuildTree(questions)
	root := models.KeypointStatisticsResp{
		Child: res,
	}
	// 把all question id 添加到树上
	services.AddQuestionIDs(&root, allQuestionCateMap)
	// 子节点count累加到父节点上
	services.DfsQuestionCount(&root)
	// 新增全部 tab
	services.AddAllTab(&root, allQuestionCateMap)
	// 排序 按试题的年月日创建时间排序
	services.SortQuestion(&root)
	return root.Child
}

func (k KeypointCountJob) deal(examCategory, examChild, jobTag, provence string) {
	filter := bson.M{"exam_category": examCategory}
	if examChild != "" {
		filter["exam_child_category"] = examChild
	}
	if jobTag != "" {
		filter["job_tag"] = jobTag
	}
	if provence != "" {
		filter["provence"] = bson.M{"provence": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(provence)}}}
	}
	filter["status"] = 5
	filter["question_real"] = 1
	// real
	realResp := k.build(filter)
	filter["question_real"] = 0
	notReal := k.build(filter)

	var realK models.KeypointStatistic
	f := bson.M{
		"exam_category":       examCategory,
		"exam_child_category": examChild,
		"job_tag":             jobTag,
		"provence":            provence,
		"question_real":       1,
	}
	_ = k.DB().Collection(models.KeypointStatisticTable).Where(f).Take(&realK)
	if realK.Id.IsZero() {
		realKeypoint := models.KeypointStatistic{
			ExamCategory:      examCategory,
			ExamChildCategory: examChild,
			JobTag:            jobTag,
			Provence:          provence,
			QuestionReal:      1,
			Keypoint:          realResp,
		}
		_, _ = k.DB().Collection(models.KeypointStatisticTable).Create(&realKeypoint)
	} else {
		realK.Keypoint = realResp
		realK.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
		_, _ = k.DB().Collection(models.KeypointStatisticTable).Where(bson.M{"_id": realK.Id}).Update(&realK)
	}

	var notRealK models.KeypointStatistic
	f["question_real"] = 0
	_ = k.DB().Collection(models.KeypointStatisticTable).Where(f).Take(&notRealK)
	if notRealK.Id.IsZero() {
		notRealKeypoint := models.KeypointStatistic{
			ExamCategory:      examCategory,
			ExamChildCategory: examChild,
			JobTag:            jobTag,
			Provence:          provence,
			QuestionReal:      0,
			Keypoint:          notReal,
		}
		_, _ = k.DB().Collection(models.KeypointStatisticTable).Create(&notRealKeypoint)
	} else {
		notRealK.Keypoint = notReal
		notRealK.UpdatedTime = time.Now().Format("2006-01-02 15:04:05")
		_, _ = k.DB().Collection(models.KeypointStatisticTable).Where(bson.M{"_id": notRealK.Id}).Update(&notRealK)
	}
	return
}

func (k KeypointCountJob) dealProvence(examCategory models.ExamCategory) {
	var wg sync.WaitGroup
	for _, area := range areas {
		wg.Add(1)
		go func(area string) {
			defer wg.Done()
			k.dealCommon(examCategory, "", area)
		}(area)
	}
	wg.Wait()
}

func (k KeypointCountJob) dealJobTag(examCategory models.ExamCategory) {
	var wg sync.WaitGroup
	for _, tag := range jobTag {
		wg.Add(1)
		go func(tag string) {
			defer wg.Done()
			k.dealCommon(examCategory, tag, "")
		}(tag)
	}
	wg.Wait()
}
