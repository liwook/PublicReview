package logger

import (
	"log/slog"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 初始化后，日志使用直接 slog.Info("dfsf")就行
//
//	func InitLogger(level string) error {
//		file, err := os.OpenFile("dianping.log", os.O_CREATE|os.O_APPEND, 0666)
//		if err != nil {
//			return err
//		}
//		//使用json格式
//		logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
//			AddSource: true,
//			Level:     LogLevel(level),
//		}))
//		slog.SetDefault(logger)
//		return nil
//	}
//
// 日志选项结构体
type LogSetting struct {
	Filename   string
	Level      string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

var LogLevel = new(slog.LevelVar)

func InitLogger(logConfig *LogSetting) {
	log := lumberjack.Logger{
		Filename:   logConfig.Filename,   //日志文件的位置
		MaxSize:    logConfig.MaxSize,    //文件最大尺寸(以mb为单位)
		MaxBackups: logConfig.MaxBackups, //保留的最大文件个数
		MaxAge:     logConfig.MaxAge,     //保留旧文件的最大天数
		LocalTime:  true,                 //使用本地时间创建时间戳
	}

	LogLevel.Set(GetLogLevel(logConfig.Level)) //这样就可以在运行时更新日志等级

	//使用json格式
	logger := slog.New(slog.NewJSONHandler(&log, &slog.HandlerOptions{
		AddSource: true,
		Level:     LogLevel,
	}))

	slog.SetDefault(logger)
}
func SetLevel(level string) {
	LogLevel.Set(GetLogLevel(level))
}

// 获得日志等级
func GetLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	return slog.LevelInfo
}
