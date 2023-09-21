package log

import (
	"fmt"
	"strings"
	"time"
)

type TextFormatter struct {
}

func (t *TextFormatter) Format(param *LoggingFormatParam) string {
	// todo 颜色输出
	now := time.Now()
	fieldsStr := ""
	if param.LoggerFields != nil {
		var sb strings.Builder
		for k, v := range param.LoggerFields {
			fmt.Fprintf(&sb, "%s=%v ", k, v)
		}
		fieldsStr = sb.String()
	}
	return fmt.Sprintf("[cob] %v | level=%s | msg=%v %s \n",
		now.Format("2006-01-02 15:04:05"), param.Level, param.Msg, fieldsStr)
}
