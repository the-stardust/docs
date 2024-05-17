package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"interview/models"
	"net/url"
)

type Manager struct {
	Middleware
}

func (sf *Manager) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		managerId := c.GetHeader("X-User-Id")
		mName, err := url.QueryUnescape(c.GetHeader("X-User-Name"))
		if err != nil {
			sf.SLogger().Error(err)
		}
		manager := models.Manager{}
		err = sf.DB().Collection("managers").Where(bson.M{"manager_id": managerId}).Take(&manager)
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				if managerId != "" {
					manager.ManagerId = managerId
					manager.ManagerName = mName
					_, err = sf.DB().Collection("managers").Create(&manager)
					if err != nil {
						sf.SLogger().Error(err)
					}
				}
			} else {
				sf.SLogger().Error(err)
			}
		} else {
			if manager.ManagerName != mName && mName != "" {
				manager.ManagerName = mName
				err = sf.DB().Collection("managers").Save(&manager)
				if err != nil {
					sf.SLogger().Error(err)
				}
			}
		}
		c.Set("user_id", managerId)
		c.Set("manager_name", mName)
		// 处理请求
		c.Next()
	}
}
