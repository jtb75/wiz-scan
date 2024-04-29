package utilities

import (
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func LogInit(level logrus.Level) {
	Log.SetLevel(level)
}
