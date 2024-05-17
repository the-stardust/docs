package database

import (
	"fmt"
	"interview/common/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MobileInfo 用户信息
type MobileInfo struct {
	GUID       string
	MobileID   string
	MobileMask string
}

// 定义一个初始化数据库的函数
func init() {
	m := global.CONFIG.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True", m.Username, m.Password, m.Path, m.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	global.Mysql = db

}
