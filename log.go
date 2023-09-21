package cob

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	greenBg   = "\033[97;42m"
	whiteBg   = "\033[90;47m"
	yellowBg  = "\033[90;43m"
	redBg     = "\033[97;41m"
	blueBg    = "\033[97;44m"
	magentaBg = "\033[97;45m"
	cyanBg    = "\033[97;46m"
	green     = "\033[32m"
	white     = "\033[37m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	blue      = "\033[34m"
	magenta   = "\033[35m"
	cyan      = "\033[36m"
	reset     = "\033[0m"
)

type LoggerConfig struct {
	Formatter LoggerFormatter
	out       io.Writer
}

type LoggerFormatter = func(params *LogFormatterParams) string

var DefaultWriter io.Writer = os.Stdout

var defaultFormatter = func(params *LogFormatterParams) string {
	var statusCodeColor = params.StatusCodeColor()
	var reset = params.ResetColor()
	if params.Latency > time.Minute {
		//超过分钟，转为秒
		params.Latency = params.Latency.Truncate(time.Second)
	}
	//todo 颜色输出
	return fmt.Sprintf("[cob] %v |%s %3d %s| %13v | %15s |%-7s %#v",
		params.TimeStamp.Format("2006-01-02 15:04:05"),
		statusCodeColor, params.StatusCode,
		reset, params.Latency, params.ClientIP,
		params.Method, params.Path)
}

type LogFormatterParams struct {
	Request        *http.Request
	TimeStamp      time.Time
	StatusCode     int
	Latency        time.Duration
	ClientIP       net.IP
	Method         string
	Path           string
	isDisplayColor bool //日志是否有颜色
}

func (l *LogFormatterParams) StatusCodeColor() string {
	switch l.StatusCode {
	case http.StatusOK:
		return green
	default:
		return red
	}
}

func (l *LogFormatterParams) ResetColor() string {
	return reset
}

func LoggingWithConfig(config LoggerConfig, next HandleFunc) HandleFunc {
	formatter := config.Formatter
	if formatter == nil {
		formatter = defaultFormatter
	}
	out := config.out
	if out == nil {
		out = DefaultWriter
	}

	return func(ctx *Context) {
		r := ctx.Request
		params := &LogFormatterParams{Request: r}
		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery
		next(ctx)
		end := time.Now()
		latency := end.Sub(start)
		ip, _, _ := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
		clientIp := net.ParseIP(ip)
		method := r.Method
		code := ctx.StatusCode

		if raw != "" {
			path = path + "?" + raw
		}

		params.TimeStamp = end
		params.Method = method
		params.Path = path
		params.ClientIP = clientIp
		params.StatusCode = code
		params.Latency = latency
		fmt.Fprintf(out, formatter(params))
	}
}

func Logging(next HandleFunc) HandleFunc {
	return LoggingWithConfig(LoggerConfig{}, next)
}
