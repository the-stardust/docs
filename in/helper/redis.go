package helper

import (
	"github.com/garyburd/redigo/redis"
	"interview/common/global"
)

func rdbPool(dbName ...string) *redis.Pool {
	return global.REDISPOOL
}

func rdbPoolBank(dbName ...string) *redis.Pool {
	return global.REDISPOOLBank
}

func RedisDel(key []string) error {
	conn := rdbPool().Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func RedisGet(key string) (string, error) {
	conn := rdbPool().Get()
	defer conn.Close()

	str, err := redis.String(conn.Do("GET", key))

	if err != nil && err.Error() == "redigo: nil returned" {
		err = nil
	}

	return str, err
}

func RedisSet(key string, value interface{}, expire ...int) error {
	conn := rdbPool().Get()
	defer conn.Close()
	var err error
	expireTime := 0
	if len(expire) > 0 {
		expireTime = expire[0]
		_, err = conn.Do("SET", key, value, "EX", expireTime)
	} else {
		_, err = conn.Do("SET", key, value)
	}

	return err
}

func RedisTTL(key string) (int, error) {
	conn := rdbPool().Get()
	defer conn.Close()

	// -2 没有这个值 -1永久
	ttl, err := redis.Int(conn.Do("TTL", key))

	if err != nil {
		ttl = -2
	}

	return ttl, err
}

func RedisSetNx(key string, expire int) (bool, error) {
	conn := rdbPool().Get()
	defer conn.Close()

	// 成功 返回 OK，nil
	// 失败 返回 空, redigo: nil returned
	_, err := redis.String(conn.Do("SET", key, 1, "EX", expire, "NX"))
	if err != nil && err.Error() == "redigo: nil returned" {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func RedisHGet(key, field string) (string, error) {
	conn := rdbPool().Get()
	defer conn.Close()

	str, err := redis.String(conn.Do("HGET", key, field))
	if err != nil && err.Error() == "redigo: nil returned" {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return str, nil
}

func RedisSIsMember(key, member string) bool {
	conn := rdbPool().Get()
	defer conn.Close()
	isHave, _ := redis.Bool(conn.Do("SISMEMBER", key, member))

	return isHave
}

func BankRedisHGet(key, field string) (string, error) {
	conn := rdbPoolBank().Get()
	defer conn.Close()

	str, err := redis.String(conn.Do("HGET", key, field))
	if err != nil && err.Error() == "redigo: nil returned" {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return str, nil
}

func RedisHSet(redisKey, key string, field []byte) error {
	conn := rdbPool().Get()
	defer conn.Close()

	_, err := conn.Do("HSET", redisKey, key, field)
	if err != nil {
		return err
	}
	return nil
}

func RedisHGetAll(key string) (map[string]string, error) {
	conn := rdbPool().Get()
	defer conn.Close()

	str, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil && err.Error() == "redigo: nil returned" {
		return make(map[string]string), nil
	}
	if err != nil {
		return make(map[string]string), err
	}
	return str, nil
}

func RedisHMSet(key string, fields map[string]interface{}) error {
	conn := rdbPool().Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(fields)...)
	return err
}

func RedisEXPIRE(key string, expire int) error {
	conn := rdbPool().Get()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expire)
	if err != nil {
		return err
	}
	return nil
}

// RedisSADD values: [redisKey, value1, value2, value3, ...]
func RedisSADD(values ...any) error {
	conn := rdbPool().Get()
	defer conn.Close()
	_, err := conn.Do("SADD", values...)
	return err
}

func RedisSREM(key, member string) error {
	conn := rdbPool().Get()
	defer conn.Close()
	_, err := conn.Do("SREM", key, member)
	return err
}

func RedisSMembers(key string) ([]string, error) {
	conn := rdbPool().Get()
	defer conn.Close()
	members, err := redis.Strings(conn.Do("SMEMBERS", key))
	return members, err
}

func WxAccessTokenTTL() int {
	ttl, _ := RedisTTL("interview:wx_access_token")
	return ttl
}
