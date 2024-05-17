package manager

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"interview/common"
	"interview/controllers"
	"interview/database"
	"interview/models"
)

type MorningReadTag struct {
	controllers.Controller
}

type listResponse struct {
	Name  string `json:"name"`
	Cover string `json:"cover"`
}

var defaultReadTag = []models.MorningReadTag{
	{Name: "热点", Cover: ""},
	{Name: "金句", Cover: ""},
	{Name: "素材积累", Cover: ""},
}

func (m *MorningReadTag) getCollection() *database.MongoWork {
	return m.DB().Collection(models.MorningReadTagTable)
}

func (m *MorningReadTag) List(c *gin.Context) {
	list := make([]models.MorningReadTag, 0)
	f := bson.M{"is_deleted": 0}
	err := m.getCollection().Where(f).Sort("-created_time").Find(&list)
	if err != nil {
		m.Error(common.CodeServerBusy, c)
		return
	}
	data := make(map[string]interface{})
	if len(list) < len(defaultReadTag) {
		list = defaultReadTag
	}
	resp := make([]listResponse, 0, len(list))
	for _, tag := range list {
		resp = append(resp, listResponse{Name: tag.Name, Cover: tag.Cover})
	}
	data["list"] = resp
	m.Success(data, c)
}

func (m *MorningReadTag) EditTag(c *gin.Context) {
	//check
	p := c.GetHeader("X-Permission")
	if p != "edit-tag" {
		m.Error(common.PermissionDenied, c, "no X-permission")
		return
	}
	var err error
	var param struct {
		Name  string `json:"name"`
		Cover string `json:"cover"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		m.Error(common.CodeInvalidParam, c)
		return
	}
	var info models.MorningReadTag
	f := bson.M{"is_deleted": 0, "name": param.Name}
	err = m.getCollection().Where(f).Take(&info)

	if errors.Is(err, mongo.ErrNoDocuments) {
		info.Name = param.Name
		info.Cover = param.Cover
		_, err = m.getCollection().Create(&info)

	} else {
		info.Cover = param.Cover
		err = m.getCollection().Save(&info)
	}
	if err != nil {
		m.Error(common.CodeServerBusy, c, err.Error())
		return
	}
	m.Success(true, c)
}

func (m *MorningReadTag) DelTag(c *gin.Context) {
	p := c.GetHeader("X-Permission")
	if p != "edit-tag" {
		m.Error(common.PermissionDenied, c, "no X-permission")
		return
	}
	var err error
	var param struct {
		Name string `json:"name"`
	}
	err = c.ShouldBindBodyWith(&param, binding.JSON)
	if err != nil {
		m.Error(common.CodeInvalidParam, c)
		return
	}
	var info models.MorningReadTag
	f := bson.M{"is_deleted": 0, "name": param.Name}
	err = m.getCollection().Where(f).Take(&info)

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		m.Error(common.CodeServerBusy, c, err.Error())
		return

	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		m.Success(true, c)
		return
	}
	info.IsDeleted = 1
	_ = m.getCollection().Save(&info)

	m.Success(true, c)
}
