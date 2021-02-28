package activitylogger

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// Options activity log options to control path, filters etc.
type Options struct {
	Methods      string
	Disabled     bool
	OutputFolder string
}

// ActivityLogger write command and request information to file for auditing purposes. Safe to call concurrently
type ActivityLogger struct {
	mu        sync.Mutex
	w         io.Writer
	contextID string
	options   Options
	fullName  string
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// createContextID return new random context id
func createContextID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// NewActivityLogger creates a new activity logger using given options
func NewActivityLogger(options Options) (*ActivityLogger, error) {
	var f *os.File
	var err error

	if options.OutputFolder != "" {
		if err := os.MkdirAll(options.OutputFolder, os.ModePerm); err != nil {
			return nil, err
		}
	}
	logPath := path.Join(options.OutputFolder, "c8y.activitylog."+time.Now().Format("2006-01-02")+".json")
	if !options.Disabled {
		f, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	if err != nil {
		return nil, err
	}
	return &ActivityLogger{
		w:         f,
		options:   options,
		fullName:  logPath,
		contextID: createContextID(8),
	}, nil
}

// GetPath returns the path to the activity log
func (l *ActivityLogger) GetPath() string {
	return l.fullName
}

// LogCommand writes the c8y cli command used to the activity log
func (l *ActivityLogger) LogCommand(cmd *cobra.Command, args []string, cmdStr string, messages ...string) {
	if l.options.Disabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	argc, _ := json.Marshal(os.Args[1:])
	if len(messages) > 0 {
		l.w.Write([]byte(fmt.Sprintf(
			`{"time":"%s","ctx":"%s","type":"command","arguments":%s,"message":"%s"}`+"\n",
			time.Now().Format(time.RFC3339Nano),
			l.contextID,
			argc,
			strings.Join(messages, ". "),
		)))
	} else {
		l.w.Write([]byte(fmt.Sprintf(
			`{"time":"%s","ctx":"%s","type":"command","arguments":%s}`+"\n",
			time.Now().Format(time.RFC3339Nano),
			l.contextID,
			argc,
		)))
	}
}

// LogCustom writes a custom entry to the activity log
func (l *ActivityLogger) LogCustom(message string) {
	if l.options.Disabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	l.w.Write([]byte(fmt.Sprintf(
		`{"time":"%s","ctx":"%s","type":"user","message":"%s"}`+"\n",
		time.Now().Format(time.RFC3339Nano),
		l.contextID,
		message,
	)))
}

// LogRequest writes a http response to the activity log
func (l *ActivityLogger) LogRequest(resp *http.Response, body *gjson.Result, responseTime int64) {
	if l.options.Disabled {
		return
	}

	if resp == nil {
		return
	}

	if !strings.Contains(l.options.Methods, resp.Request.Method) {
		return
	}

	query, err := url.QueryUnescape(resp.Request.URL.RawQuery)
	if err != nil {
		query = resp.Request.URL.RawQuery
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.w.Write([]byte(fmt.Sprintf(
		`{"time":"%s","ctx":"%s","type":"request","method":"%s","host":"%s","path":"%s","query":"%s","accept":"%s","processingMode":"%s","statusCode":%d,"responseTimeMS":%d,"responseSelf":"%s"}`+"\n",
		time.Now().Format(time.RFC3339Nano),
		l.contextID,
		resp.Request.Method,
		resp.Request.URL.Host,
		resp.Request.URL.Path,
		query,
		resp.Request.Header.Get("Accept"),
		resp.Request.Header.Get("X-Cumulocity-Processing-Mode"),
		resp.StatusCode,
		responseTime,
		body.Get("self").Str,
	)))
}
