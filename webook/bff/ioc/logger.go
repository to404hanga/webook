package ioc

import (
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger_OldVersion() logger.Logger {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func InitLogger() logger.Logger {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "E:/VSCode/GO/webook/webook/log/user.log", // 指定日志文件路径
		MaxSize:    50,                                        // 每个日志文件的最大大小，单位：MB
		MaxBackups: 3,                                         // 保留旧的日志文件最大个数
		MaxAge:     28,                                        // 保留旧的日志文件最大天数
		Compress:   true,                                      // 是否压缩旧的日志文件
	}

	// 创建 zap 日志核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zapcore.DebugLevel, // 设置日志级别
	)

	l := zap.New(core, zap.AddCaller())
	res := logger.NewZapLogger(l)

	return res
}
