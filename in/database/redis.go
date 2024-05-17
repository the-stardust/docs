/*
 * @Descripttion:redis服务
 * @Author: yangxiaoyang
 * @Date: 2020-12-09 16:39:13
 */
package database

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"interview/common/global"
)

func init() {
	c := global.CONFIG.Redis
	pool := &redis.Pool{ //实例化一个连接池
		MaxIdle:     c.MaxIdle,     //最初的连接数量
		MaxActive:   c.MaxActive,   //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: c.IdleTimeout, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			setdb := redis.DialDatabase(c.DB)
			setPasswd := redis.DialPassword(c.Password)
			d, err := redis.Dial("tcp", c.Addr, setdb, setPasswd)
			if err != nil {
				panic(fmt.Errorf("Fatal error redis:%v\n", err))
			}
			if c.Password != "" {
				d.Do("AUTH", c.Password)
			}
			d.Do("expire", c.Expire)
			global.REDISDB = d
			return d, err
		},
	}
	global.REDISPOOL = pool

	pool = &redis.Pool{ //实例化一个连接池
		MaxIdle:     c.MaxIdle,     //最初的连接数量
		MaxActive:   c.MaxActive,   //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: c.IdleTimeout, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库
			setdb := redis.DialDatabase(c.BankDB)
			setPasswd := redis.DialPassword(c.Password)
			d, err := redis.Dial("tcp", c.Addr, setdb, setPasswd)
			if err != nil {
				panic(fmt.Errorf("Fatal error redis:%v\n", err))
			}
			if c.Password != "" {
				d.Do("AUTH", c.Password)
			}
			d.Do("expire", c.Expire)
			return d, err
		},
	}
	global.REDISPOOLBank = pool
}
