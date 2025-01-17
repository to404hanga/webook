package startup

import "github.com/to404hanga/pkg404/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
