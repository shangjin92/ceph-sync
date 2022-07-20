package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
	normal = 0
	light  = 4
)

// getColor return color by log level
func getColor(level logrus.Level) int {
	var levelColor int
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	case logrus.InfoLevel:
		levelColor = blue
	default:
		levelColor = blue
	}
	return levelColor
}

func ColorLevelText(levelText string, level logrus.Level) string {
	return colorText(levelText, getColor(level))
}

func HighlightText(text string) string {
	return colorText(text, light)
}

func NormalText(text string) string {
	return colorText(text, normal)
}

func colorText(text string, textColor int) string {
	return fmt.Sprintf("\x1B[%dm%s\x1B[0m", textColor, text)
}
