package models

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"interview/common"
	"interview/common/global"
	"interview/common/rediskey"
	"interview/database"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/remeh/sizedwaitgroup"
	"go.mongodb.org/mongo-driver/bson"
)

type Member struct {
	UserId        string `bson:"user_id" json:"user_id" redis:"user_id"`
	Name          string `bson:"name" json:"name" redis:"name"`                            //成员名称
	IdentityType  int8   `bson:"identity_type" json:"identity_type" redis:"identity_type"` //身份类型 0管理员 1老师2学生
	Status        int8   `bson:"status" json:"status" redis:"status"`                      //身份类型 0正常 2已移除
	IdentityName  string `bson:"-" json:"identity_name"`                                   //身份类型 0管理员 1老师2学生
	Avatar        string `bson:"-" json:"avatar" redis:"avatar"`                           //头像
	Nickname      string `bson:"-" json:"-" redis:"nickname"`
	IsAllowAnswer int8   `bson:"-" json:"is_allow_answer" redis:"-"`
	Idx           int    `bson:"-" json:"-" redis:"-"`
	IsCached      int8   `bson:"-" json:"-" redis:"-"`
}
type InterviewClass struct {
	DefaultField         `bson:",inline"`
	Name                 string   `bson:"name" json:"name"`                           // 班级名称
	CreatorUserId        string   `bson:"creator_user_id" json:"creator_user_id"`     //创建者
	CreatorUserName      string   `bson:"creator_user_name" json:"creator_user_name"` //创建者
	Members              []Member `bson:"members" json:"members"`                     //成员
	MemberCount          int      `bson:"member_count" json:"member_count"`
	IsTeacherCQ          int8     `bson:"is_teacher_cq" json:"is_teacher_cq"`                       //老师创建试题 0不可以 1可以
	IsStudentCQ          int8     `bson:"is_student_cq" json:"is_student_cq"`                       //学生创建试题 0不可以 1可以
	IsStudentSeeAnswer   int8     `bson:"is_student_see_answer" json:"is_student_see_answer"`       //学生看别人回答 0不可以 1可以
	IsStudentSeeComment  int8     `bson:"is_student_see_comment" json:"is_student_see_comment"`     //学生看别人评论 0不可以 1可以
	IsAnswerTimeOutAlert int8     `bson:"is_answer_time_out_alert" json:"is_answer_time_out_alert"` //答题超时提醒
	StructuredGroup      int8     `bson:"structured_group" json:"structured_group"`                 //结构化小组模式
	AnswerTimeOut        int64    `bson:"answer_time_out" json:"answer_time_out"`                   //答题超时时间(仅作为提醒)
	ClassCode            int64    `bson:"class_code" json:"class_code"`                             //班级码用于进入班级
	ClassTeacherCode     int64    `bson:"-" json:"class_teacher_code"`                              //老师码用于进入班级
	Status               int8     `bson:"status" json:"status"`                                     // 状态，5正常,9已关闭（学员不可见）
	UserName             string   `bson:"-" json:"user_name"`                                       //创建者
	IdentityType         int8     `bson:"-" json:"identity_type"`                                   //身份类型 1老师2学生
	MemberStatus         int8     `bson:"-" json:"member_status"`                                   //0正常 2已移除
	TeacherCount         int      `bson:"-" json:"teacher_count"`                                   //老师数量
	StudentCount         int      `bson:"-" json:"student_count"`                                   //学生数量
	IsDeleted            int8     `json:"is_deleted" bson:"is_deleted"`                             // 班级是否被删除，0否，1是
	CurriculaId          string   `json:"curricula_id" bson:"curricula_id"`
}

type InterviewClassAndClassCount struct {
	InterviewClass
	QuestionCount  int64  `json:"question_count"`
	CurriculaTitle string `json:"curricula_title"`
}

const (
	interviewClassTable = "interview_class"
)

var InterviewClassModel = &InterviewClass{}

func (i *InterviewClass) getCollection() *database.MongoWork {
	return i.DB().Collection(interviewClassTable)
}

func (i *InterviewClass) UserIsBelongToCurricula(userID, curriculaID string) bool {
	if userID == "" || curriculaID == "" {
		return false
	}
	f := bson.M{"is_deleted": 0, "status": 5, "curricula_id": curriculaID, "members.user_id": userID}
	var info InterviewClass
	err := i.getCollection().Where(f).Take(&info)
	if err != nil {
		return false
	}
	if info.Id.IsZero() {
		return false
	}
	return true
}

func (i *InterviewClass) GetClassesMap(classIDs []string) (map[string]InterviewClass, error) {
	if len(classIDs) == 0 {
		return nil, fmt.Errorf("class_ids is empty")
	}
	classIDMap := make(map[string]bool)
	for _, classID := range classIDs {
		classIDMap[classID] = true
	}
	objectIDs := make([]primitive.ObjectID, 0, len(classIDMap))
	for classID := range classIDMap {
		objectIDs = append(objectIDs, i.ObjectID(classID))
	}
	var res []InterviewClass
	err := i.getCollection().Where(bson.M{"is_deleted": 0, "_id": bson.M{"$in": objectIDs}}).Find(&res)
	if err != nil {
		return nil, err
	}
	var resMap = make(map[string]InterviewClass, len(res))
	for _, class := range res {
		resMap[class.Id.Hex()] = class
	}
	return resMap, nil
}

// GetUserClass 获取用户所属班级
func (i *InterviewClass) GetUserClass(userID string) (*InterviewClass, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is empty")
	}
	var info InterviewClass
	f := bson.M{"is_deleted": 0, "members": bson.M{"$elemMatch": bson.M{"status": 0, "user_id": userID}}}
	err := i.getCollection().Where(f).Sort("-created_time").Take(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetUsersClassMap 获取用户们所属班级map 班级管理员也可以
func (i *InterviewClass) GetUsersClassMap(userIDs []string) (map[string]InterviewClass, error) {
	var list []InterviewClass
	f := bson.M{"is_deleted": 0, "members": bson.M{"$elemMatch": bson.M{"status": 0, "user_id": bson.M{"$in": userIDs}}}}
	err := i.getCollection().Where(f).Find(&list)
	if err != nil {
		return nil, err
	}
	res := make(map[string]InterviewClass, len(list))
	for _, class := range list {
		for _, member := range class.Members {
			res[member.UserId] = class
		}
	}
	return res, nil
}

// GetUserCurriculaIDs 获取用户所属的考试ids
func (i *InterviewClass) GetUserCurriculaIDs(userID string) []string {
	filter := bson.M{"is_deleted": 0, "status": 5, "members": bson.M{"$elemMatch": bson.M{"user_id": userID, "status": 0}}}
	var list []InterviewClass
	var currIDs []string
	// 查询自己所属的班级所属的考试id interview_class.curricula_id
	err := i.getCollection().Where(filter).Find(&list)
	if err != nil {
		return currIDs
	}
	// 需要去重
	res := make([]string, 0, len(list))
	for _, class := range list {
		res = append(res, class.CurriculaId)
	}
	return res
}

func (sf *InterviewClass) GetMembersInfo(classId string, userIds []string, appCode string, retryTimes int) map[string]Member {
	var err error
	//获取 用户头像昵称
	var unCacheUids = []string{}
	membersChan := make(chan Member, len(userIds))
	memberMapChan := make(chan map[string]Member, 1)

	go func() {
		mp := map[string]Member{}
		for v := range membersChan {
			if v.IsCached == 0 {
				unCacheUids = append(unCacheUids, v.UserId)
			} else {
				mp[v.UserId] = v
			}
		}
		memberMapChan <- mp

	}()
	sw := sizedwaitgroup.New(20)
	for i, v := range userIds {
		sw.Add()
		go func(idx int, cId string, uId string, swg *sizedwaitgroup.SizedWaitGroup) {
			defer swg.Done()
			member := sf.getRedisMemberInfo(cId, uId)
			if member != nil && (member.Name != "" || member.Avatar != "") {
				member.Idx = idx
				member.IsCached = 1
				membersChan <- *member
			} else {
				t := Member{UserId: uId, Idx: idx}
				membersChan <- t
			}
		}(i, classId, v, &sw)
	}
	sw.Wait()
	close(membersChan)
	memberInfoMap := <-memberMapChan
	close(memberMapChan)
	unCacheMember := map[string]Member{}
	var class InterviewClass
	err = sf.DB().Collection("interview_class").Where(bson.M{"_id": sf.ObjectID(classId)}).Take(&class)
	if err == nil {
		for _, v := range class.Members {
			if common.InArr(v.UserId, unCacheUids) {
				unCacheMember[v.UserId] = v
			}
		}
	} else {
		sf.SLogger().Error(err)
		return memberInfoMap
	}
	if len(unCacheUids) > 0 {
		type GatewayParam struct {
			Ids     []string `json:"ids"`
			AppCode string   `json:"appCode"`
		}
		pm := GatewayParam{
			Ids:     unCacheUids,
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
					memberInfoMap[v.Guid] = Member{UserId: v.Guid, Name: unCacheMember[v.Guid].Name}
					identityType := 2
					if vv, ok := unCacheMember[v.Guid]; ok {
						identityType = int(vv.IdentityType)
					}
					sf.SetRedisMemberInfo(classId, v.Guid, []interface{}{"name", unCacheMember[v.Guid].Name, "nick_name", v.Nickname, "user_id", v.Guid, "avatar", v.Avatar, "identity_type", identityType}...)
				}
			}
		} else {
			sf.SLogger().Error(err)
			if retryTimes > 0 {
				return sf.GetMembersInfo(classId, unCacheUids, appCode, retryTimes-1)
			}
		}

	}
	return memberInfoMap
}
func (sf *InterviewClass) SetRedisMemberInfo(classId string, uid string, fv ...interface{}) error {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	key := fmt.Sprintf("%s:%s:%s", rediskey.InterviewMemberInfo, classId, uid)
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

// 删除map
func (sf *InterviewClass) DeleteRedisMemberInfo(classId string, uid string) error {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	key := fmt.Sprintf("%s:%s:%s", rediskey.InterviewMemberInfo, classId, uid)
	_, err = rdb.Do("expire", key, 0)
	if err != nil {
		sf.SLogger().Error(err)
		return err
	}
	return nil
}

func (sf *InterviewClass) getRedisMemberInfo(classId string, uid string) *Member {
	var err error
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	key := fmt.Sprintf("%s:%s:%s", rediskey.InterviewMemberInfo, classId, uid)
	res, err := redis.Values(rdb.Do("HGETALL", key))
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	if res == nil {
		return nil
	}
	member := Member{}
	err = redis.ScanStruct(res, &member)
	if err != nil {
		sf.SLogger().Error(err)
		return nil
	}
	return &member
}

type InterviewQuestion struct {
	DefaultField             `bson:",inline"`
	ClassId                  string   `json:"class_id" bson:"class_id"`
	Name                     string   `bson:"name" json:"name"`                                                 // 试题名称
	Desc                     string   `bson:"desc" json:"desc"`                                                 // 试题描述
	CreatorUserId            string   `bson:"creator_user_id" json:"creator_user_id"`                           //创建者
	CreatorUserIdentity      string   `bson:"creator_user_identity" json:"creator_user_identity"`               //创建者身份
	IsTeacherComment         int8     `bson:"is_teacher_comment" json:"is_teacher_comment"`                     //老师点评 0不可以 1可以
	IsStudentComment         int8     `bson:"is_student_comment" json:"is_student_comment"`                     //学生点评 0不可以 1可以
	IsTeacherCanChangeStatus int8     `bson:"is_teacher_can_change_status" json:"is_teacher_can_change_status"` // 试题是否能被老师修改，默认是只允许创建者修改
	AnswerStyle              int8     `bson:"answer_style" json:"answer_style"`                                 //0默认都可以看见试题 1.排队模式只有老师放行学生 学生才能看到题
	AllowAnswerUserIds       []string `bson:"allow_answer_user_ids" json:"allow_answer_user_ids"`               //排队模式下 允许看见题的用户id
	Status                   int32    `bson:"status" json:"status"`                                             // 状态0未上架 5正常 9已删除（不可见）
	UserName                 string   `bson:"-" json:"user_name"`
	IsAllowAnswer            int8     `bson:"-" json:"is_allow_answer"`       //0不允许 1允许
	SourceType               int8     `json:"source_type" bson:"source_type"` // 题目来源 0是老师或者学生上传 1是管理员上传
}

func (sf *InterviewQuestion) GetQuestionName(id string) string {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	s, err := redis.String(rdb.Do("HGET", rediskey.InterviewQuestionId2Name, id))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			var question InterviewQuestion
			err = sf.DB().Collection("interview_questions").Where(bson.M{"_id": sf.ObjectID(id)}).Take(&question)
			if err == nil {
				rdb.Do("HSET", rediskey.InterviewQuestionId2Name, id, question.Name)
				s = question.Name
			} else {
				sf.SLogger().Error("interview_questions 无此id" + id)
			}

		} else {
			sf.SLogger().Error(err)
		}
	}
	return s
}

type VoiceContent struct {
	Start int64  `bson:"start" json:"start"`
	End   int64  `bson:"end" json:"end"`
	Text  string `bson:"text" json:"text"`
}

//	type Answer struct {
//		VoiceLength  float64        `bson:"voice_length" json:"voice_length"`   //语音长度
//		VoiceUrl     string         `bson:"voice_url" json:"voice_url"`         //语音url
//		VoiceContent []VoiceContent `bson:"voice_content" json:"voice_content"` //每段内容
//		VoiceText    string         `bson:"voice_text" json:"voice_text"`       //语音文本
//		Status       int32          `bson:"status" json:"status"`
//	}
//
//	type AnswerLog struct {
//		DefaultField `bson:",inline"`
//		UserId       string   `bson:"user_id" json:"user_id"`
//		ClassId      string   `json:"class_id" bson:"class_id"`
//		QuestionId   string   `bson:"question_id" json:"question_id"`
//		Answer       []Answer `bson:"answer" json:"answer"` // 试题描述
//		UserName     string   `bson:"-" json:"user_name"`
//		Avatar       string   `bson:"-" json:"avatar"`
//		QuestionName string   `bson:"-" json:"question_name"`
//		IsDeleted    int      `bson:"is_deleted" json:"is_deleted"`
//	}
type Comment struct {
	Content     string `bson:"content" json:"content"`         //内容
	Speed       string `bson:"speed" json:"speed"`             //语速
	Interaction string `bson:"interaction" json:"interaction"` //互动
	Confident   string `bson:"confident" json:"confident"`     //自信
	Grade       string `bson:"grade" json:"grade"`             //评分
	GAnswer     `bson:",inline"`
	CommentText string `bson:"comment_text" json:"comment_text"` //文本内容
	IsDeleted   int    `bson:"is_deleted" json:"is_deleted"`
}
type Reply struct {
	Star float32 `bson:"star" json:"star"` //星
}
type AnswerComment struct {
	DefaultField        `bson:",inline"`
	UserId              string  `bson:"user_id" json:"user_id"`
	ClassId             string  `json:"class_id" bson:"class_id"`
	CommentedUserId     string  `bson:"commented_user_id" json:"commented_user_id"` //被点评人
	AnswerLogId         string  `bson:"answer_log_id" json:"answer_log_id"`         //回答记录id
	QuestionId          string  `bson:"question_id" json:"question_id"`             //试题id
	StructuredGroup     int8    `bson:"structured_group" json:"structured_group"`   //结构化小组模式
	Comment             Comment `bson:"comment" json:"comment"`                     // 试题描述
	Reply               Reply   `bson:"reply" json:"reply"`                         //回复
	ReplyComment        Comment `bson:"reply_comment" json:"reply_comment"`         //回复
	UserName            string  `bson:"-" json:"user_name"`
	Avatar              string  `bson:"-" json:"avatar"`
	QuestionName        string  `bson:"-" json:"question_name"`
	CommentedUserName   string  `bson:"-" json:"commented_user_name"`
	CommentedUserAvatar string  `bson:"-" json:"commented_user_avatar"`
	SpeechTextTaskId    string  `bson:"speech_text_task_id" json:"speech_text_task_id"` // 语音文本任务ID
}
