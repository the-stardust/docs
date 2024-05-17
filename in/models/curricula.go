package models

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const CurriculaTableName = "interview_curricula"

type Curricula struct {
	DefaultField   `bson:",inline"`
	IsDelete       int8             `json:"is_delete" bson:"is_delete"`
	CurriculaTitle string           `json:"curricula_title" bson:"curricula_title"`
	CreateUserId   string           `json:"create_user_id" bson:"create_user_id"`
	Name           string           `json:"name" bson:"name"`
	Sort           int              `json:"sort" bson:"sort"`
	Status         int              `json:"status" bson:"status"`
	AdminList      []AdminListModel `json:"admin_list" bson:"admin_list"`
}

var CurriculaModel = &Curricula{}

func (sf *Curricula) TableName() string {
	return CurriculaTableName
}

type AdminListModel struct {
	AdminId     string `json:"admin_id" bson:"admin_id"`
	Type        int    `json:"type" bson:"type"` // 0创建者 1管理员
	CreatedTime string `json:"created_time" bson:"created_time"`
	Name        string `json:"name" bson:"name"`
	Avatar      string `json:"avatar" bson:"-"` //头像
}

func (c *Curricula) CheckCurriculaAdmin(userID, curriculaID string) bool {
	if userID == "" || curriculaID == "" {
		return false
	}
	var info Curricula
	err := c.DB().Collection(CurriculaTableName).Where(bson.M{"admin_list": bson.M{"$elemMatch": bson.M{"admin_id": userID}}, "_id": c.ObjectID(curriculaID)}).Take(&info)
	if err != nil {
		fmt.Println("check err:", err.Error())
		return false
	}
	return true
}

func (c *Curricula) GetUserControlCurrID(userID string) []string {
	if userID == "" {
		return []string{}
	}
	f := bson.M{"admin_list": bson.M{"$elemMatch": bson.M{"admin_id": userID}}, "is_delete": 0}
	var list []Curricula
	err := c.DB().Collection(CurriculaTableName).Where(f).Find(&list)
	if err != nil {
		return []string{}
	}
	res := make([]string, 0, len(list))
	for _, v := range list {
		res = append(res, v.Id.Hex())
	}
	return res
}

func (c *Curricula) GetCurriculaMap(id []string) map[string]Curricula {
	res := make(map[string]Curricula)
	if len(id) == 0 {
		return res
	}
	idObj := make([]primitive.ObjectID, 0, len(id))
	for _, v := range id {
		idObj = append(idObj, c.ObjectID(v))
	}
	f := bson.M{"_id": bson.M{"$in": idObj}}
	var list []Curricula
	err := c.DB().Collection(CurriculaTableName).Where(f).Find(&list)
	if err != nil {
		return res
	}
	for _, v := range list {
		res[v.Id.Hex()] = v
	}
	return res
}
