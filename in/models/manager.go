package models

import (
	"github.com/garyburd/redigo/redis"
	"go.mongodb.org/mongo-driver/bson"
	"interview/common/rediskey"
)

type Manager struct {
	DefaultField `bson:",inline"`
	ManagerName  string `bson:"manager_name"`
	ManagerId    string `bson:"manager_id"`
}

func (sf *Manager) GetManagerName(id string) string {
	rdb := sf.RDBPool().Get()
	defer rdb.Close()
	s, err := redis.String(rdb.Do("HGET", rediskey.ManagerId2Name, id))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			var manager Manager
			err = sf.DB().Collection("managers").Where(bson.M{"manager_id": id}).Take(&manager)
			if err == nil {
				rdb.Do("HSET", rediskey.ManagerId2Name, id, manager.ManagerName)
				s = manager.ManagerName
			} else {
				sf.SLogger().Error("managers 无此manager_id" + id)
			}

		} else {
			sf.SLogger().Error(err)
		}
	}
	return s
}
