/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

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
