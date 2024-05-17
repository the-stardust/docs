package models

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"math/rand"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/remeh/sizedwaitgroup"
	"go.mongodb.org/mongo-driver/bson"
)

type InterviewGPT struct {
	DefaultField `bson:",inline"`
}

type CommonContentStyle struct {
	StyleType string            `json:"style_type" bson:"style_type"` //u 下划线
	Attr      map[string]string `json:"attr" bson:"attr"`
}
type CommonContent struct {
	DataType int8                 `json:"type" bson:"type"` // 1，代表文本；2，代表图片；3，代表文本图片混合；4，下划线；5代表填空题的那个空 6 分数
	Text     string               `json:"text" bson:"text"`
	Latex    string               `json:"latex" bson:"latex"`
	Style    []CommonContentStyle `json:"style" bson:"style"`
	Image    string               `json:"image,omitempty" bson:"image,omitempty"`
	Width    string               `json:"width,omitempty" bson:"width,omitempty"`
	Height   string               `json:"height,omitempty" bson:"height,omitempty"`
	Content  []CommonContent      `json:"content,omitempty" bson:"content,omitempty"`
}

func (cc CommonContent) ContentToStr() string {
	tempStr := ""
	if cc.DataType == 1 {
		//文字
		tempStr = cc.Text
	} else if cc.DataType == 3 {
		//图文
		for _, v := range cc.Content {
			if v.DataType == 1 {
				if tempStr == "" {
					tempStr = v.Text
				} else {
					tempStr += v.Text
				}
			}
		}
	}
	return tempStr
}

type GQuestion struct {
	DefaultField          `bson:",inline"`
	Name                  string        `bson:"name" json:"name" redis:"name"`                                  // 试题名称
	NameStruct            CommonContent `bson:"name_struct" json:"name_struct" redis:"name_struct"`             // 试题名称
	NameDesc              string        `json:"name_desc" bson:"name_desc" redis:"name_desc"`                   // 漫画题的总结性内容
	Desc                  string        `bson:"desc" json:"desc" redis:"desc"`                                  // 试题描述
	Tags                  []string      `bson:"tags" json:"tags" redis:"tags"`                                  // 试题tag
	Answer                string        `bson:"answer" json:"answer" redis:"answer"`                            // 试题答案
	Thinking              string        `bson:"thinking" json:"thinking" redis:"thinking"`                      // 解题思路（第三方）
	CategoryId            string        `bson:"category_id" json:"category_id" redis:"category_id"`             // 试题分类
	CreatorUserId         string        `bson:"creator_user_id" json:"creator_user_id" redis:"creator_user_id"` //创建者
	Status                int32         `bson:"status" json:"status" redis:"status"`                            // 状态0未上架 5正常 9已删除（不可见）
	UserName              string        `bson:"-" json:"user_name" redis:"-"`
	ExamCategory          string        `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory     string        `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory      []string      `json:"question_category" bson:"question_category"`     //题分类
	Year                  int           `json:"year" bson:"year"`                               // 年份
	Month                 int           `json:"month" bson:"month"`                             // 月份
	Day                   int           `json:"day" bson:"day"`                                 // 日
	Province              string        `json:"province" bson:"province"`
	City                  string        `json:"city" bson:"city"`
	District              string        `json:"district" bson:"district"`
	GPTAnswer             []GPTAnswer   `bson:"gpt_answer" json:"gpt_answer"`           // 解题思路
	JobTag                string        `json:"job_tag" bson:"job_tag"`                 // 岗位标签，如海关、税务局等
	QuestionSource        string        `json:"question_source" bson:"question_source"` // 试题来源
	AreaCodes             []string      `json:"area_codes" bson:"area_codes"`           // 地区代码
	Moment                string        `json:"moment" bson:"moment"`                   // 上午或者下午
	Date                  string        `json:"date" bson:"-"`                          //日期
	AnswerCount           int64         `json:"answer_count" bson:"-"`                  // 答题次数
	CorrectCount          int64         `json:"correct_count" bson:"-"`                 // 批改次数
	PeopleCount           int64         `json:"people_count" bson:"people_count"`       // 答题人数
	TTSUrl                TTSUrl        `json:"tts_url" bson:"tts_url"`                 // 合成语音地址
	ManagerID             string        `json:"manager_id" bson:"manager_id"`           // 试题上传者的x-user-id
	ManagerName           string        `json:"manager_name" bson:"-"`
	GPTPreviewStatus      string        `json:"gpt_preview_status" bson:"-"`        // gpt预生成 0没有 1正在生成 2 完成
	PreviewIdeas          []string      `json:"preview_ideas" bson:"-"`             // 预生成的答题思路
	PreviewStandardAnswer []string      `json:"preview_standard_answer" bson:"-"`   // 预生成的标准回答
	AnswerTime            int64         `json:"answer_time" bson:"answer_time"`     // 答题时间，单位为秒
	QuestionReal          int8          `json:"question_real" bson:"question_real"` //是否真题, 0是模拟题，1是真题
	MyAnswerCount         int64         `json:"my_answer_count" bson:"-"`
	IsNew                 bool          `json:"is_new" bson:"-"`
	RequireAnswerDuration int           `json:"require_answer_duration" bson:"-"`                   // 规定题答题时间 秒
	QuestionContentType   int8          `json:"question_content_type" bson:"question_content_type"` // 试题类别，0普通题，1漫画题
	ScorePoint            string        `json:"score_point" bson:"score_point"`                     // 评分要点
	ExplainUrl            string        `json:"explain_url" bson:"explain_url"`                     // 试题讲解
}

// 普通题只取纯文本内容，漫画题取desc，以便GPT相关接口使用
func (sf GQuestion) GetWantedQuestionContent() string {
	tempStr := ""
	if sf.QuestionContentType == 0 {
		tempStr = sf.Name
	} else if sf.QuestionContentType == 1 {
		tempStr = sf.NameStruct.ContentToStr() + "\n" + "漫画描述：" + sf.NameDesc
	}
	return tempStr
}

func (sf *GQuestion) TableName() string {
	return "g_interview_questions"
}

type GCustomQuestion struct {
	DefaultField      `bson:",inline"`
	UserId            string    `bson:"user_id" json:"user_id" redis:"user_id"` //创建者
	Name              string    `bson:"name" json:"name" redis:"name"`          // 试题名称
	GPTAnswer         GPTAnswer `bson:"gpt_answer" json:"gpt_answer"`
	UserAvatar        string    `json:"user_avatar" bson:"-"`
	UserName          string    `json:"user_name" bson:"-"`
	AnswerType        int8      `json:"answer_type" bson:"answer_type"`                 // 1是回答公务员面试类的问题，2是随便提问，不限范围
	ExamCategory      string    `json:"exam_category" bson:"exam_category"`             // 考试分类
	ExamChildCategory string    `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory  []string  `json:"question_category" bson:"question_category"`     //题分类
}
type GQuestionCategory struct {
	DefaultField `bson:",inline"`
	Name         string `bson:"name" json:"name" redis:"name"`
	Prompt       string `bson:"prompt" json:"prompt" redis:"prompt"`
	Status       int32  `bson:"status" json:"status" redis:"status"`
}

func (sf *GQuestion) GetQuestionName(id string) string {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	s, err := redis.String(rdb.Do("HGET", rediskey.InterviewGPTQuestionId2Name, id))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			var question InterviewQuestion
			err = sf.DB().Collection("g_interview_questions").Where(bson.M{"_id": sf.ObjectID(id)}).Take(&question)
			if err == nil {
				rdb.Do("HSET", rediskey.InterviewGPTQuestionId2Name, id, question.Name)
				s = question.Name
			} else {
				sf.SLogger().Error("g_interview_questions 无此id" + id)
			}

		} else {
			sf.SLogger().Error(err)
		}
	}
	return s
}

type GTeacher struct {
	DefaultField `bson:",inline"`
	UserId       string `bson:"user_id" json:"user_id" redis:"user_id"`
	Name         string `bson:"name" json:"name" redis:"name"`                            //成员名称
	IdentityType int8   `bson:"identity_type" json:"identity_type" redis:"identity_type"` //身份类型 0管理员1老师
	Avatar       string `bson:"-" json:"avatar" redis:"avatar"`                           //头像
	Nickname     string `bson:"-" json:"-" redis:"nickname"`
	Status       int32  `bson:"status" json:"status"` // 状态 5正常 9已移除
}

type GVoiceContent struct {
	Start int64  `bson:"start" json:"start"`
	End   int64  `bson:"end" json:"end"`
	Text  string `bson:"text" json:"text"`
}

type GAnswer struct {
	VoiceLength  float64        `bson:"voice_length" json:"voice_length"`   //语音长度
	VoiceUrl     string         `bson:"voice_url" json:"voice_url"`         //语音url
	VoiceContent []VoiceContent `bson:"voice_content" json:"voice_content"` //每段内容
	VoiceText    string         `bson:"voice_text" json:"voice_text"`       //语音文本
	Status       int32          `bson:"status" json:"status"`
}
type GPTComment struct {
	Content string `bson:"content" json:"content"`
}
type GPTAnswer struct {
	Content string `bson:"content" json:"content"`
}
type GAnswerLog struct {
	DefaultField      `bson:",inline"`
	ExamCategory      string    `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string    `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	QuestionCategory  []string  `json:"question_category" bson:"question_category"`     //题分类
	UserId            string    `bson:"user_id" json:"user_id"`
	QuestionId        string    `bson:"question_id" json:"question_id"`
	Answer            []GAnswer `bson:"answer" json:"answer"` // 学员回答内容
	UserName          string    `bson:"-" json:"user_name"`
	Avatar            string    `bson:"-" json:"avatar"`
	QuestionName      string    `bson:"question_name" json:"question_name"` //todo 不保存数据库?
	IsDeleted         int       `bson:"is_deleted" json:"is_deleted"`

	GPTComment            GPTComment            `bson:"gpt_comment" json:"gpt_comment"` // gpt点评
	GPTAnswers            []GPTAnswer           `bson:"-" json:"gpt_answers"`           // 解题思路
	GPTAnswer             GPTAnswer             `bson:"-" json:"gpt_answer"`            // 解题思路
	GPTStandardAnswer     string                `json:"gpt_standard_answer" bson:"-"`
	SID                   string                `json:"sid" bson:"sid"`                     // 科大讯飞sid
	ThoughtIndex          int8                  `json:"thought_index" bson:"thought_index"` // 用户上次浏览的思路索引
	PracticeType          int8                  `json:"practice_type" bson:"practice_type"` // 练习模式种类，0是全部、11是看题-普通模式，12是看题-对镜模式，13是看题-考官模式 21是听题-普通模式，22是听题-对镜模式，23是听题-考官模式
	Province              string                `json:"province" bson:"province"`
	City                  string                `json:"city" bson:"city"`
	District              string                `json:"district" bson:"district"`
	ReviewId              string                `json:"review_id" bson:"review_id"`             // 测评ID
	ReviewLogId           string                `json:"review_log_id" bson:"review_log_id"`     // 测评logID
	CorrectStatus         int8                  `json:"correct_status" bson:"correct_status"`   // 1已点评
	TeacherCorrectContent TeacherCorrectComment `bson:"teacher_correct" json:"teacher_correct"` // 老师点评
	ScoreType             int                   `bson:"-" json:"score_type"`                    // 打分方式 1仅GPT 2仅老师 3先GPT后老师
	ClassId               string                `json:"class_id" bson:"class_id"`
	LogType               int8                  `json:"log_type" bson:"log_type"`                           // 1是面试AI，2是面试云
	QuestionContentType   int8                  `json:"question_content_type" bson:"question_content_type"` // 试题类别，0普通题，1漫画题
	NameStruct            CommonContent         `bson:"name_struct" json:"name_struct" redis:"name_struct"` // 试题名称
	JobTag                string                `bson:"-" json:"job_tag"`                                   // 岗位标签
	IsTeacherComment      bool                  `json:"is_teacher_comment" bson:"-"`
	SpeechTextTaskId      string                `bson:"speech_text_task_id" json:"speech_text_task_id"` // 语音文本任务ID
}

func (g *GAnswerLog) TableName() string {
	return "g_interview_answer_logs"
}

type AggregateGAnswerLog struct {
	All   GAnswerLog `bson:"all"`
	Count int        `bson:"count"`
}
type AggregateGAnswerLogRes struct {
	GAnswerLog
	Count                int  `json:"count" bson:"-"`
	IsTeacherComment     bool `json:"is_teacher_comment" bson:"-"`
	IsTeacherCommentUser bool `json:"is_teacher_comment_user" bson:"-"`
}

type AggregateGAnswerLogListRes struct {
	Total int64                    `json:"total"`
	List  []AggregateGAnswerLogRes `json:"list"`
}

type TeacherCorrectComment struct {
	Content     string `bson:"content" json:"content"`           //内容
	Speed       string `bson:"speed" json:"speed"`               //语速
	Interaction string `bson:"interaction" json:"interaction"`   //互动
	Confident   string `bson:"confident" json:"confident"`       //自信
	Grade       string `bson:"grade" json:"grade"`               //评分
	CommentText string `bson:"comment_text" json:"comment_text"` //文本内容
	GAnswer
}

type GUser struct {
	UserId       string `bson:"user_id" json:"user_id" redis:"user_id"`
	IdentityType int8   `bson:"identity_type" json:"identity_type" redis:"identity_type"` //身份类型 0管理员1老师2学生
	Avatar       string `bson:"-" json:"avatar" redis:"avatar"`                           //头像
	Nickname     string `bson:"-" json:"nick_name" redis:"nick_name"`
}

func (sf *InterviewGPT) GetUsersInfo(userIds []string, appCode string, retryTimes int) map[string]GUser {
	//获取 用户头像昵称
	tempUserIds := []string{}
	for _, v := range userIds {
		if !common.InArr(v, tempUserIds) {
			tempUserIds = append(tempUserIds, v)
		}
	}
	userIds = tempUserIds
	usersChan := make(chan GUser, len(userIds))
	userMapChan := make(chan map[string]GUser, 1)
	go func() {
		mp := map[string]GUser{}
		for v := range usersChan {
			mp[v.UserId] = v
		}
		userMapChan <- mp

	}()
	sw := sizedwaitgroup.New(20)
	for _, v := range userIds {
		sw.Add()
		go func(uId string, swg *sizedwaitgroup.SizedWaitGroup) {
			defer swg.Done()
			user := sf.getRedisUserInfo(uId)
			if user != nil {
				usersChan <- *user
			} else {
				t := GUser{UserId: uId, IdentityType: 2}
				usersChan <- t
			}
		}(v, &sw)
	}
	sw.Wait()
	close(usersChan)
	userInfoMap := <-userMapChan
	close(userMapChan)
	if len(userIds) > 0 && appCode != "" {
		type GatewayParam struct {
			Ids     []string `json:"ids"`
			AppCode string   `json:"appCode"`
		}
		pm := GatewayParam{
			Ids:     userIds,
			AppCode: appCode}
		res, err := common.HttpPostJson(global.CONFIG.ServiceUrls.UserLoginUrl+"/new-userinfo/users-info", pm)
		type TempRes struct {
			Msg  string                     `json:"msg"`
			Code string                     `json:"code"`
			Data map[string]GatewayUserInfo `json:"data"`
		}
		if err == nil {
			r := TempRes{}
			err = json.Unmarshal(res, &r)
			if err != nil {
				sf.SLogger().Error(err)
			} else {
				for _, v := range r.Data {
					if user, ok := userInfoMap[v.Guid]; ok {
						if user.Nickname != v.Nickname || user.Avatar != v.Avatar {
							sf.SetRedisUserInfo(v.Guid, []interface{}{"nick_name", v.Nickname, "user_id", v.Guid, "avatar", v.Avatar, "identity_type", userInfoMap[v.Guid].IdentityType}...)
						}

					}
				}
			}
		} else {
			sf.SLogger().Error(err)
			if retryTimes > 0 {
				return sf.GetUsersInfo(userIds, appCode, retryTimes-1)
			}
		}

	}
	return userInfoMap
}
func (sf *InterviewGPT) SetRedisUserInfo(uid string, fv ...interface{}) error {
	var err error
	rdb := sf.RDBPool().Get()
	key := fmt.Sprintf("%s:%s", rediskey.InterviewGPTUserInfo, uid)
	fieldsValues := []interface{}{key}
	fieldsValues = append(fieldsValues, fv...)
	_, err = rdb.Do("HMSET", fieldsValues...)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(100000)
	rdb.Do("expire", key, 86400*30+random)
	return nil
}
func (sf *InterviewGPT) getRedisUserInfo(uid string) *GUser {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	key := fmt.Sprintf("%s:%s", rediskey.InterviewGPTUserInfo, uid)
	exists, _ := redis.Bool(rdb.Do("EXISTS", key))
	if !exists {
		return nil
	}
	res, err := redis.Values(rdb.Do("HGETALL", key))
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	if res == nil {
		return nil
	}
	user := GUser{}
	err = redis.ScanStruct(res, &user)
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	return &user
}

type GPTAnswerSet struct {
	SystemContent string  `bson:"system_content" json:"system_content" redis:"system_content"`
	PromptPrefix  string  `bson:"prompt_prefix" json:"prompt_prefix" redis:"prompt_prefix"`
	Temperature   float32 `bson:"temperature" json:"temperature" redis:"temperature"`
	TopP          float32 `bson:"top_p" json:"top_p" redis:"top_p"`
}

func (sf *InterviewGPT) SetRedisGPTSet(key string, fv ...interface{}) error {
	var err error
	rdb := sf.RDBPool().Get()
	fieldsValues := []interface{}{key}
	fieldsValues = append(fieldsValues, fv...)
	_, err = rdb.Do("HMSET", fieldsValues...)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	return nil
}
func (sf *InterviewGPT) GetRedisGPTSet(key string) *GPTAnswerSet {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	exists, _ := redis.Bool(rdb.Do("EXISTS", key))
	if !exists {
		return nil
	}
	res, err := redis.Values(rdb.Do("HGETALL", key))
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	if res == nil {
		return nil
	}
	answer := GPTAnswerSet{}
	err = redis.ScanStruct(res, &answer)
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	return &answer
}

type DataInfo struct {
	DataTitle     string   `json:"data_title" bson:"data_title"`
	DataIDs       []string `json:"data_ids" bson:"data_ids"`
	Url           string   `json:"url" bson:"url"`
	RequestMethod string   `json:"request_method" bson:"request_method"`
	RequestBody   string   `json:"request_body" bson:"request_body"`
}
type UserFeedback struct {
	DefaultField `bson:",inline"`
	UserId       string   `json:"user_id" bson:"user_id" mongo-index:"1"`
	DataInfo     DataInfo `json:"data_info" bson:"data_info"`     // 反馈的页面相关信息
	FastRemark   []string `json:"fast_remark" bson:"fast_remark"` // 快捷反馈列表
	Remark       string   `json:"remark" bson:"remark"`           // 用户反馈信息
	SourceType   string   `json:"source_type" bson:"source_type"`
	ImageList    []string `json:"image_list" bson:"image_list"`

	NickName   string `json:"nick_name" bson:"-"`
	Avatar     string `json:"avatar" bson:"-"`
	MobileID   string `json:"mobile_id" bson:"-"`
	MobileMask string `json:"mobile_mask" bson:"-"`
}

type CategoryGPTPrompt struct {
	DefaultField      `bson:",inline"`
	AnswerType        int8     `json:"answer_type" bson:"answer_type"`                 // 类别，0是面试题目+学员回答进行点评，1是面试题生成思路，2是学员自己提问,3是后台生成标准回答时的提示
	ExamCategory      string   `json:"exam_category" bson:"exam_category"`             //考试分类
	ExamChildCategory string   `json:"exam_child_category" bson:"exam_child_category"` //考试子分类
	SystemContent     string   `json:"system_content" bson:"system_content"`           // GPT系统设定信息
	Prompt            string   `json:"prompt" bson:"prompt"`                           // GPT提示内容
	Temperature       float64  `json:"temperature" bson:"temperature"`                 //温度
	TopP              float64  `json:"top_p" bson:"top_p"`
	QuestionCategory  []string `json:"question_category" bson:"question_category"`
}

type GIDMap struct {
	DefaultField `bson:",inline"`
	LongID       string `json:"long_id" bson:"long_id"`
	ShortID      string `json:"short_id" bson:"short_id"`
}

type TTSUrl struct {
	MaleVoiceUrl      string  `json:"male_voice_url" bson:"male_voice_url"`
	FemaleVoiceUrl    string  `json:"female_voice_url" bson:"female_voice_url"`
	MaleVoiceLength   float64 `json:"male_voice_length" bson:"male_voice_length"`
	FemaleVoiceLength float64 `json:"female_voice_length" bson:"female_voice_length"`
}

type TTSPrompt struct {
	DefaultField     `bson:",inline"`
	MaleHeadPrompt   string  `json:"male_head_prompt" bson:"male_head_prompt" redis:"male_head_prompt"`
	MaleTailPrompt   string  `json:"male_tail_prompt" bson:"male_tail_prompt" redis:"male_tail_prompt"`
	FeMaleHeadPrompt string  `json:"female_head_prompt" bson:"female_head_prompt" redis:"female_head_prompt"`
	FeMaleTailPrompt string  `json:"female_tail_prompt" bson:"female_tail_prompt" redis:"female_tail_prompt"`
	PromptType       int8    `json:"prompt_type" bson:"prompt_type" redis:"prompt_type"`
	HeadVoiceLength  float64 `json:"head_voice_length" bson:"head_voice_length" redis:"head_voice_length"`
	TailVoiceLength  float64 `json:"tail_voice_length" bson:"tail_voice_length" redis:"tail_voice_length"`
}

type QuestionHash struct {
	Hash             string   `json:"hash"`
	QuestionIds      []string `json:"question_ids"`
	QuestionIdsTotal int      `json:"question_ids_total"`
}

type QuestionSimilar struct {
	Hash        string   `json:"hash"`
	QuestionIds []string `json:"question_ids"`
	SimilarType int      `json:"similar_type"` // 1完全相同 2相似
}

func (g *GAnswerLog) GetUserCategoryAnswerCount(examCategory, examChildCategory, uid string, questionReal int) map[string][]string {
	filter := bson.M{"exam_category": examCategory}
	if examChildCategory != "" {
		filter["exam_child_category"] = examChildCategory
	}
	filter["user_id"] = uid
	filter["log_type"] = 1
	filter["question_category.0"] = bson.M{"$exists": true}
	var logs []GAnswerLog
	countMap := make(map[string][]string)

	err := g.DB().Collection(g.TableName()).Where(filter).Sort("+created_time").Find(&logs)
	if err != nil {
		return countMap
	}
	qIDs := make([]primitive.ObjectID, 0, len(logs))
	for _, log := range logs {
		qIDs = append(qIDs, g.ObjectID(log.QuestionId))
	}
	// 查询 question_real
	var finialLogs []GQuestion
	f := bson.M{"_id": bson.M{"$in": qIDs}, "question_real": questionReal}
	err = g.DB().Collection("g_interview_questions").Where(f).Find(&finialLogs)
	if err != nil {
		return countMap
	}
	// make map
	finQIDs := make(map[string]bool)
	for _, log := range finialLogs {
		finQIDs[log.Id.Hex()] = true
	}
	qidMap := make(map[string]bool)
	for _, v := range logs {
		// 不符合questionReal的过滤
		if _, ok := finQIDs[v.QuestionId]; !ok {
			continue
		}
		// 多次练习只计算一次
		if _, ok := qidMap[v.QuestionId]; ok {
			continue
		}
		qidMap[v.QuestionId] = true
		k := strings.Join(v.QuestionCategory, "")
		if _, ok := countMap[k]; !ok {
			countMap[k] = []string{v.QuestionId}
		} else {
			countMap[k] = append(countMap[k], v.QuestionId)
		}
	}
	return countMap
}
