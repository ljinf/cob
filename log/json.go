package log

import (
	"encoding/json"
	"fmt"
	"time"
)

type JsonFormatter struct {
}

func (j *JsonFormatter) Format(param *LoggingFormatParam) string {
	if param.LoggerFields == nil {
		param.LoggerFields = make(Fields)
	}
	param.LoggerFields["time"] = time.Now().Format("2006-01-02 15:04:05")
	param.LoggerFields["msg"] = param.Msg
	param.LoggerFields["level"] = param.Level.Level()
	marshal, err := json.Marshal(param.LoggerFields)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", string(marshal))
}
