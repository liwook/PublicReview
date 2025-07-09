package config

import (
	"log/slog"
	"review/pkg/logger"
	"review/pkg/mail"
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
	MailOption   *mail.MailSetting
)

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
		// reloadAllSection() //更新所有数据
		//只更新日志的数据
		err := ReadSection("log", &LogOption)
		if err != nil {
			slog.Error("read log section error", "err", err)
		}
		//查看是否有更新了日志等级
		if logger.GetLogLevel(LogOption.Level) != logger.LogLevel.Level() {
			logger.SetLevel(LogOption.Level)
		}
	})
	return nil
}

// 分段读取
func ReadSection(key string, v any) error {
	return viper.UnmarshalKey(key, v)
}

// 用于重新读取配置
func reloadAllSection() {
	err := ReadSection("server", &ServerOption)
	if err != nil {
		slog.Error("read server section error", "err", err)
	}
	err = ReadSection("mysql", &MysqlOption)
	if err != nil {
		slog.Error("read mysql section error", "err", err)
	}
	err = ReadSection("log", &LogOption)
	if err != nil {
		slog.Error("read log section error", "err", err)
	}
}

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
	err = ReadSection("mail", &MailOption)
	if err != nil {
		panic(err)
	}
}
