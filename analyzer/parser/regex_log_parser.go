package parser

import (
	"errors"
	"github.com/yihongzhi/log-kit/collector/sender"
	"github.com/yihongzhi/log-kit/config"
	"regexp"
	"time"
)

type RegexLogParser struct {
	regx       *regexp.Regexp
	timeFormat string
}

func NewRegexLogParser(config *config.LogParserConfig) *RegexLogParser {
	return &RegexLogParser{
		regx:       regexp.MustCompile("(?ms)" + config.Pattern),
		timeFormat: config.TimeFormat,
	}
}

func (p *RegexLogParser) Parse(logMessage *sender.LogMessage) (*LogContent, error) {
	strings := p.regx.FindStringSubmatch(logMessage.Content)
	if strings == nil {
		log.Warnf("log message has no match: appId=%s,content=%s", logMessage.AppId, logMessage.Content)
		return nil, errors.New("log message has no match")
	}
	timeStr := p.matchedValue(strings, "time")
	timeValue, err := time.Parse(p.timeFormat, timeStr)
	if err != nil {
		log.Warnln("parse log time error", err)
		return nil, errors.New("parse log time error")
	}
	return &LogContent{
		AppId:     logMessage.AppId,
		Host:      logMessage.Host,
		ParseTime: time.Now(),
		Time:      timeValue,
		Level:     p.matchedValue(strings, "level"),
		TxId:      p.matchedValue(strings, "tx_id"),
		SpanId:    p.matchedValue(strings, "span_id"),
		Field: map[string]string{
			"thread": p.matchedValue(strings, "thread"),
			"method": p.matchedValue(strings, "method"),
		},
		Content: p.matchedValue(strings, "content"),
	}, nil
}

func (p *RegexLogParser) matchedValue(strings []string, field string) string {
	index := p.regx.SubexpIndex(field)
	if index == -1 {
		return ""
	}
	return strings[index]
}
