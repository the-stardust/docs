package config

import (
	"fmt"
	"interview/common/global"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 配置文件路径 (包名 + 配置文件名 )
const defaultConfigFile = "./config/config.yaml"
const jobConfigFile = "../../config/config.yaml"

// 初始化配置文件
func init() {
	v := viper.New()
	env := os.Getenv("GO_ENV")
	if env == "" {
		path, _ := os.Getwd()
		if strings.Contains(path, "job") {
			v.SetConfigFile(jobConfigFile)
		} else {
			v.SetConfigFile(defaultConfigFile)
		}
	} else {
		v.SetConfigFile("/root/conf/" + env + "_config.yaml")
	}
	// 读取配置文件中的配置信息，并将信息保存 到 v中
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Fatal error config file:%+v\n", err))
	}
	// 监控配置文件
	v.WatchConfig()
	// 配置文件改变，则将 v中的配置信息，刷新到 global.CONFIG
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(&global.CONFIG); err != nil {
			panic(fmt.Sprintf("Fatal error config file:%v\n", err))
		}
	})
	// 将 v 中的配置信息 反序列化成 结构体 (将v 中配置信息 刷新到 global.CONFIG)
	if err := v.Unmarshal(&global.CONFIG); err != nil {
		panic(fmt.Sprintf("Fatal error config file:%v\n", err))
	}
	// 保存 viper 实例 v
	global.VIPER = v
}
