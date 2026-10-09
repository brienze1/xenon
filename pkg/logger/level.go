package logger

import (
	"os"
	"strings"
)

type level int

const (
	levelTrace level = iota
	levelDebug
	levelInfo
	levelWarn
	levelError
	levelOff
)

var levelNames = map[level]string{
	levelTrace: "TRACE",
	levelDebug: "DEBUG",
	levelInfo:  "INFO",
	levelWarn:  "WARN",
	levelError: "ERROR",
	levelOff:   "OFF",
}

func (l level) String() string {
	return levelNames[l]
}

func configuredLevel() level {
	configured := strings.ToUpper(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	for candidate, name := range levelNames {
		if name == configured {
			return candidate
		}
	}
	return levelInfo
}

func shouldPrintLog(candidate level) bool {
	threshold := configuredLevel()
	return threshold != levelOff && candidate >= threshold
}
