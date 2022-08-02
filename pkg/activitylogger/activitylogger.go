package activitylogger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/fileutilities"
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
		if err := fileutilities.CreateDirs(options.OutputFolder); err != nil {
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

type CommandEntry struct {
	Time      string   `json:"time,omitempty"`
	Context   string   `json:"ctx,omitempty"`
	Type      string   `json:"type,omitempty"`
	Arguments []string `json:"arguments,omitempty"`
	Message   string   `json:"message,omitempty"`
}

type RequestEntry struct {
	Time           string `json:"time,omitempty"`
	Context        string `json:"ctx,omitempty"`
	Type           string `json:"type,omitempty"`
	Method         string `json:"method,omitempty"`
	Host           string `json:"host,omitempty"`
	Path           string `json:"path,omitempty"`
	Query          string `json:"query,omitempty"`
	Accept         string `json:"accept,omitempty"`
	ProcessingMode string `json:"processingMode,omitempty"`
	StatusCode     int    `json:"statusCode,omitempty"`
	ResponseTimeMS int    `json:"responseTimeMS,omitempty"`
	ResponseSelf   string `json:"responseSelf,omitempty"`
}

// LogCommand writes the c8y cli command used to the activity log
func (l *ActivityLogger) LogCommand(cmd *cobra.Command, args []string, cmdStr string, messages ...string) {
	if l.options.Disabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	argc, _ := json.Marshal(os.Args[1:])
	if len(messages) > 0 && messages[0] != "" {
		fmt.Fprintf(l.w,
			`{"time":"%s","ctx":"%s","type":"command","arguments":%s,"message":"%s"}`+"\n",
			time.Now().Format(time.RFC3339Nano),
			l.contextID,
			argc,
			strings.Join(messages, ". "),
		)
	} else {
		fmt.Fprintf(l.w,
			`{"time":"%s","ctx":"%s","type":"command","arguments":%s}`+"\n",
			time.Now().Format(time.RFC3339Nano),
			l.contextID,
			argc,
		)
	}
}

// LogCustom writes a custom entry to the activity log
func (l *ActivityLogger) LogCustom(message string) {
	if l.options.Disabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.w,
		`{"time":"%s","ctx":"%s","type":"user","message":"%s"}`+"\n",
		time.Now().Format(time.RFC3339Nano),
		l.contextID,
		message,
	)
}

// LogRequest writes a http response to the activity log
func (l *ActivityLogger) LogRequest(resp *http.Response, body gjson.Result, responseTime int64) {
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
	query = strings.ReplaceAll(query, `\`, `\\`)
	if err != nil {
		query = resp.Request.URL.RawQuery
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		cacheTag := resp.Header.Get("ETag")
		isCachedResponse := cacheTag != ""
		fmt.Fprintf(l.w,
			`{"time":"%s","ctx":"%s","type":"request","method":"%s","host":"%s","path":"%s","query":"%s","accept":"%s","processingMode":"%s","statusCode":%d,"responseTimeMS":%d,"responseSelf":"%s","etag":"%s","cached":%v}`+"\n",
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
			cacheTag,
			isCachedResponse,
		)
	} else {
		errorResponse := body.Raw
		if !strings.HasPrefix(errorResponse, "{") {
			errorResponse = "\"" + errorResponse + "\""
		}
		fmt.Fprintf(l.w,
			`{"time":"%s","ctx":"%s","type":"request","method":"%s","host":"%s","path":"%s","query":"%s","accept":"%s","processingMode":"%s","statusCode":%d,"responseTimeMS":%d,"responseSelf":"%s","responseError":%s}`+"\n",
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
			errorResponse,
		)
	}
}

type Filter struct {
	Host     string `json:"host"`
	Type     string `json:"type"`
	DateFrom string `json:"dateFrom"`
	DateTo   string `json:"dateTo"`
}

// GetLogEntries get log entries
func (l *ActivityLogger) GetLogEntries(filter Filter, onValue func(line []byte) error) error {
	path := l.GetPath()
	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	entry := &RequestEntry{}
	for scan.Scan() {
		line := scan.Bytes()
		err := json.Unmarshal(line, entry)
		if err != nil {
			log.Fatalf("unmarshal error. line=%s, err=%s", line, err)
			continue
		}
		if filter.Type != "" && !strings.EqualFold(filter.Type, entry.Type) {
			continue
		}
		if filter.Host != "" && entry.Host != filter.Host {
			continue
		}
		if filter.DateTo != "" && entry.Time > filter.DateTo {
			continue
		}
		if filter.DateFrom != "" && entry.Time < filter.DateFrom {
			continue
		}
		if err := onValue(line); err != nil {
			return err
		}
	}
	return nil
}
