package config

import (
	"dianping/pkg/logger"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	ServerOption *ServerSetting
	MysqlOption  *MysqlSetting
	LogOption    *logger.LogSetting
	RedisOption  *RedisSetting
	JwtOption    *JWTSetting
)

var sections = make(map[string]any)

type ServerSetting struct {
	RunMode      string
	HttpPort     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type MysqlSetting struct {
	UserName     string
	Password     string
	Host         string
	DbName       string
	MaxIdleConns int
	MaxOpenConns int
}

type RedisSetting struct {
	Host     string
	Password string
	PoolSize int
}

type JWTSetting struct {
	Secret string
	Issuer string
	Expire time.Duration
}

// func init() {
// 	InitConfig()
// }

// viper的使用
// 打开配置文件进行读取
//
//	func ReadConfigFile() error {
//		//viper是可以开箱即用的，这样写法就类似单例模式
//		//也可以创建viper 比如 vp:=viper.New()
//		viper.SetConfigFile("../configs/config.yaml") // 指定配置文件名和位置
//		return viper.ReadInConfig()
//	}

func ReadConfigFile(path string) error {
	//viper是可以开箱即用的，这样写法就类似单例模式
	//也可以创建viper 比如 vp:=viper.New()
	viper.SetConfigFile(path) // 指定配置文件名和位置
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.WatchConfig() //该函数内部是开启了一个新协程去监听配置文件是否更新
	//设置回调函数
	viper.OnConfigChange(func(in fsnotify.Event) {
		reloadAllSection()
		//查看是否有更新了日志等级
		level := viper.GetString("log.level")
		// fmt.Println("new_level:", level)
		logger.LogLevel.Set(logger.GetLogLevel(level))
	})
	return nil
}

// 分段读取
func ReadSection(key string, v any) error {
	// return viper.UnmarshalKey(key, v)

	err := viper.UnmarshalKey(key, v)
	if err != nil {
		return nil
	}

	//增加读取section的存储记录，以便在重新加载配置的方法中进行处理
	if _, ok := sections[key]; !ok {
		sections[key] = v
	}
	return nil
}

// 用于重新读取配置
func reloadAllSection() error {
	for k, v := range sections {
		if err := ReadSection(k, v); err != nil {
			return nil
		}
	}
	return nil
}

// func InitConfig() {
// 	if err := ReadConfigFile(); err != nil {
// 		panic(err)
// 	}

// 	err := ReadSection("server", &ServerOption)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = ReadSection("mysql", &MysqlOption)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = ReadSection("log", &LogOption)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func InitConfig(path string) {
	if err := ReadConfigFile(path); err != nil {
		panic(err)
	}

	err := ReadSection("server", &ServerOption)
	if err != nil {
		panic(err)
	}
	err = ReadSection("mysql", &MysqlOption)
	if err != nil {
		panic(err)
	}
	err = ReadSection("log", &LogOption)
	if err != nil {
		panic(err)
	}

	err = ReadSection("redis", &RedisOption)
	if err != nil {
		panic(err)
	}

	err = ReadSection("jwt", &JwtOption)
	if err != nil {
		panic(err)
	}
}
