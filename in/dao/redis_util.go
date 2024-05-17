package dao

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

// redis  公共部分 start=============>>>>>>>>>>>
func closeRedisConnect(conn redis.Conn) error {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// redis  公共部分   end=============>>>>>>>>>>>

// redis  string 相关  start=============>>>>>>>>>>>

func (r RedisDao) setExRedisDao(key, value string, expireTime int64) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := rdb.Do("SETEX", key, expireTime, value)
	if err != nil {
		msg := fmt.Sprintf("SETEX key=%s value=%s expireTime=%d err=%s", key, value, expireTime, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}

func (r RedisDao) setRedisDao(key, value string) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := rdb.Do("SET", key, value)
	if err != nil {
		msg := fmt.Sprintf("SET key=%s value=%s err=%s", key, value, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}

func (r RedisDao) getRedisDao(key string) (string, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	value, err := redis.String(rdb.Do("GET", key))
	if err != nil {
		msg := fmt.Sprintf("GET key=%s err=%s", key, err.Error())
		r.SLogger().Error(msg)
		return value, err
	}
	return value, nil
}

func (r RedisDao) mGetRedisDao(keys []interface{}) ([]string, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	value, err := redis.Strings(rdb.Do("MGET", keys...))
	if err != nil {
		msg := fmt.Sprintf("GET key=%v err=%s", keys, err.Error())
		r.SLogger().Error(msg)
		return value, err
	}
	return value, nil
}

func (r RedisDao) existsRedisDao(key string) (bool, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	flag, err := redis.Bool(rdb.Do("EXISTS", key))
	if err != nil {
		msg := fmt.Sprintf("EXISTS key=%s err=%s", key, err.Error())
		r.SLogger().Error(msg)
		return flag, err
	}
	return flag, nil
}

func (r RedisDao) hKeysRedisDao(key string) ([]string, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	values, err := redis.Strings(rdb.Do("HKEYS", key))
	if err != nil {
		msg := fmt.Sprintf("HKEYS key=%s err=%s", key, err.Error())
		r.SLogger().Error(msg)
		return values, err
	}
	return values, nil
}

func (r RedisDao) deleteRedisDao(key string) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := redis.Bool(rdb.Do("DEL", key))
	if err != nil {
		msg := fmt.Sprintf("DEL key=%s err=%s", key, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}

// redis  string 相关  end=============>>>>>>>>>>>

// redis  map 相关  start=============>>>>>>>>>>>

func (r RedisDao) hExistsRedisDao(key, field string) (bool, error) {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	flag, err := redis.Bool(rdb.Do("HEXISTS", key, field))
	if err != nil {
		msg := fmt.Sprintf("HEXISTS key=%s field=%s err=%s", key, field, err.Error())
		r.SLogger().Error(msg)
		return flag, err
	}
	return flag, nil
}

func (r RedisDao) hSetRedisDao(key, field string, value interface{}) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := rdb.Do("HSET", key, field, value)
	if err != nil {
		msg := fmt.Sprintf("HSET key=%s field=%s value=%v err=%s", key, field, value, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}

func (r RedisDao) mHSetRedisDao(key string, valueMap map[string]interface{}) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := rdb.Do("MHSET", redis.Args{}.Add(key).AddFlat(
		valueMap)...)
	if err != nil {
		msg := fmt.Sprintf("MHSET key=%s  value=%v err=%s", key, valueMap, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}
func (r RedisDao) HDelRedisDao(key string, field []string) error {
	rdb := r.RDBPool().Get()
	defer func() {
		err := closeRedisConnect(rdb)
		if err != nil {
			r.SLogger().Error("close redis connect err", err.Error())
		}
	}()
	_, err := rdb.Do("HDEL", redis.Args{}.Add(key).AddFlat(
		field)...)
	if err != nil {
		msg := fmt.Sprintf("HDEL key=%s  field=%v err=%s", key, field, err.Error())
		r.SLogger().Error(msg)
		return err
	}
	return nil
}

// redis  map 相关  end=============>>>>>>>>>>>
