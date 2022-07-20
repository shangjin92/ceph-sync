package logger

import (
	"bytes"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type LogFormatter struct {
	WithColor bool
	ShowLine  bool
}

// Format yyyy-mm-dd HH:MM:SS [level] message
func (s *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logTime := entry.Time.Format("2006-01-02 15:04:05")
	levelText := strings.ToUpper(entry.Level.String())
	msg := bytes.NewBufferString(fmt.Sprintf("%s", logTime))
	if s.WithColor {
		levelText = ColorLevelText(levelText, entry.Level)
	}
	msg.WriteString(fmt.Sprintf(" [%s]", levelText))
	if s.ShowLine {
		_, file := path.Split(entry.Caller.File)
		msg.WriteString(fmt.Sprintf("%s:%d", file, entry.Caller.Line))
	}
	msg.WriteString(fmt.Sprintf(" %s\n", entry.Message))
	return msg.Bytes(), nil
}

func newLfsHook(logName string, maxTime, maxRemainTime time.Duration) (logrus.Hook, error) {
	logsWriter, err := rotatelogs.New(
		logName+".%Y%m%d%H",
		rotatelogs.WithRotationTime(maxTime),
		rotatelogs.WithMaxAge(maxRemainTime),
	)

	if err != nil {
		return nil, fmt.Errorf("config local file system for logger error: %v", err)
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: logsWriter,
		logrus.InfoLevel:  logsWriter,
		logrus.WarnLevel:  logsWriter,
		logrus.ErrorLevel: logsWriter,
		logrus.FatalLevel: logsWriter,
		logrus.PanicLevel: logsWriter,
	}, &LogFormatter{})

	return lfsHook, nil
}

func SetLogFileHook(logName string, maxTime, maxRemainTime time.Duration) error {
	hook, err := newLfsHook(logName, maxTime, maxRemainTime)
	if err != nil {
		return err
	}
	logrus.AddHook(hook)
	return nil
}

func SetLogStdHook(isDebug bool) {
	logrus.SetOutput(ioutil.Discard)
	levels := []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
	if isDebug {
		levels = append(levels, logrus.DebugLevel)
	}
	logrus.AddHook(&writer.Hook{
		Writer:    os.Stdout,
		LogLevels: levels,
	})
}
