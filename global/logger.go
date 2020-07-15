package global

import (
	"github.com/jwdev42/logger"
	"os"
)

const Default_Loglevel = logger.Level_Error

var log = logger.New(os.Stdout, Default_Loglevel, " - ")

func GetLogger() *logger.Logger {
	return log
}
