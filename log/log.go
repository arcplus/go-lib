package log

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/arcplus/go-lib/pool"
	"github.com/rs/zerolog"
)

// Level indicate log level.
type Level = zerolog.Level
type Sampler = zerolog.Sampler

// BasicSampler is a sampler that will send every Nth events, regardless of
// there level.
type BasicSampler = zerolog.BasicSampler

// Level
const (
	DebugLevel = zerolog.DebugLevel
	InfoLevel  = zerolog.InfoLevel
	WarnLevel  = zerolog.WarnLevel
	ErrorLevel = zerolog.ErrorLevel
	FatalLevel = zerolog.FatalLevel
	Disabled   = zerolog.Disabled
)

// Log struct.
type Log struct {
	mu           *sync.RWMutex
	zl           *zerolog.Logger
	sampler      zerolog.Sampler
	depth        int
	callerEnable bool
	stackEnable  bool
	kv           []interface{} // must be even len
}

var zl = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

var logger = Log{
	mu: &sync.RWMutex{},
	zl: &zl,
}

// prefixSize is used internally to trim the user specific path from the
// front of the returned filenames from the runtime call stack.
var prefixSize int

func init() {
	zerolog.MessageFieldName = "msg"
	zerolog.TimestampFieldName = "ts"
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	_, file, _, ok := runtime.Caller(0)
	if file == "?" {
		return
	}
	if ok {
		size := len(file)
		suffix := len("github.com/arcplus/go-lib/log/log.go")
		prefixSize = len(strings.TrimSuffix(file[:size-suffix], "vendor/")) // remove vendor
	}
}

// Logger return copy of default logger
func Logger() Log {
	l := logger
	l.kv = make([]interface{}, 0, 8)
	return l
}

// SetOutput set multi log writer, careful, all SetXXX method are non-thread safe.
func SetOutput(w ...io.Writer) {
	switch len(w) {
	case 0:
		return
	case 1:
		zl = zl.Output(w[0])
	default:
		zl = zl.Output(zerolog.MultiLevelWriter(w...))
	}
}

// SetLevel set global log max level.
func SetLevel(l Level) {
	zerolog.SetGlobalLevel(l)
}

var depth = 2

// SetCallDepth set call depth for show line number.
func SetCallDepth(n int) {
	depth = n
}

// SetAttachment add global kv to logger
func SetAttachment(kv map[string]interface{}) {
	if len(kv) == 0 {
		return
	}
	ctx := zl.With()
	for k, v := range kv {
		switch vv := v.(type) {
		case string:
			ctx = ctx.Str(k, vv)
		case float64:
			ctx = ctx.Float64(k, vv)
		case int64:
			ctx = ctx.Int64(k, vv)
		case int:
			ctx = ctx.Int(k, vv)
		default:
			ctx = ctx.Interface(k, vv)
		}
	}
	zl = ctx.Logger()
}

func Debug(v string) {
	l := logger
	l.depth++
	l.Debug(v)
}

func Debugf(format string, v ...interface{}) {
	l := logger
	l.depth++
	l.Debugf(format, v...)
}

func DebugEnabled() bool {
	return zl.Debug().Enabled()
}

func Info(v string) {
	l := logger
	l.depth++
	l.Info(v)
}

func Infof(format string, v ...interface{}) {
	l := logger
	l.depth++
	l.Infof(format, v...)
}

func Warn(v string) {
	l := logger
	l.depth++
	l.Warn(v)
}

func Warnf(format string, v ...interface{}) {
	l := logger
	l.depth++
	l.Warnf(format, v...)
}

func Error(v string) {
	l := logger
	l.depth++
	l.Error(v)
}

func Errorf(format string, v ...interface{}) {
	l := logger
	l.depth++
	l.Errorf(format, v...)
}

func Fatal(v string) {
	l := logger
	l.depth++
	l.Fatal(v)
}

func Fatalf(format string, v ...interface{}) {
	l := logger
	l.depth++
	l.Fatalf(format, v...)
}

func KV(k string, v interface{}) Log {
	l := logger
	l.kv = append(l.kv, k, v)
	return l
}

func (l Log) KV(k string, v interface{}) Log {
	l.kv = append(l.kv, k, v)
	return l
}

func KVPair(kv map[string]interface{}) Log {
	l := logger
	for k, v := range kv {
		l.kv = append(l.kv, k, v)
	}
	return l
}

// SetKV change kv slice
func (l *Log) SetKV(k string, v interface{}) Log {
	l.mu.Lock()
	l.kv = append(l.kv, k, v)
	l.mu.Unlock()
	return *l
}

func Trace(v string) Log {
	l := logger
	l.kv = append(l.kv, "tid", v)
	return l
}

func (l Log) Trace(v string) Log {
	l.kv = append(l.kv, "tid", v)
	return l
}

func Caller() Log {
	l := logger
	l.callerEnable = true
	return l
}

func (l Log) Caller() Log {
	l.callerEnable = true
	return l
}

func WithStack() Log {
	l := logger
	l.stackEnable = true
	return l
}

func (l Log) WithStack() Log {
	l.stackEnable = true
	return l
}

func Sample(sampler Sampler) Log {
	l := logger
	l.sampler = sampler
	return l
}

func Skip(n int) Log {
	l := logger
	l.depth += n
	return l
}

func (l Log) Skip(n int) Log {
	l.depth += n
	return l
}

func (l Log) DebugEnabled() bool {
	return l.zl.Debug().Enabled()
}

func (l Log) Debug(v string) {
	l.depth++
	l.Debugf(v)
}

func (l Log) Debugf(format string, v ...interface{}) {
	l.levelLog(DebugLevel, format, v...)
}

func (l Log) Info(v string) {
	l.depth++
	l.Infof(v)
}

func (l Log) Infof(format string, v ...interface{}) {
	l.levelLog(InfoLevel, format, v...)
}

func (l Log) Warn(v string) {
	l.depth++
	l.Warnf(v)
}

func (l Log) Warnf(format string, v ...interface{}) {
	l.levelLog(WarnLevel, format, v...)
}

func (l Log) Error(v string) {
	l.depth++
	l.Errorf(v)
}

func (l Log) Errorf(format string, v ...interface{}) {
	l.levelLog(ErrorLevel, format, v...)
}

func (l Log) Fatal(v string) {
	l.depth++
	l.Fatalf(v)
}

func (l Log) Fatalf(format string, v ...interface{}) {
	l.levelLog(FatalLevel, format, v...)
}

func (l Log) levelLog(lv Level, format string, v ...interface{}) {
	evt := l.zl.WithLevel(lv)

	if l.sampler != nil {
		s := l.zl.Sample(l.sampler)
		evt = s.WithLevel(lv)
	}

	for i, ln := 0, len(l.kv); i < ln; i = i + 2 {
		switch vv := l.kv[i+1].(type) {
		case string:
			evt.Str(l.kv[i].(string), vv)
		case float64:
			evt.Float64(l.kv[i].(string), vv)
		case int64:
			evt.Int64(l.kv[i].(string), vv)
		case int:
			evt.Int(l.kv[i].(string), vv)
		default:
			evt.Interface(l.kv[i].(string), l.kv[i+1])
		}
	}

	if l.callerEnable {
		_, file, line, ok := runtime.Caller(depth + l.depth)
		if ok {
			if prefixSize != 0 && len(file) > prefixSize {
				file = file[prefixSize:]
			}
			file += ":" + strconv.Itoa(line)
			evt.Str("caller", file)
		}
	}

	if l.stackEnable {
		evt.Str("stack", TakeStacktrace(l.depth))
	}

	evt.Msgf(format, v...)
}

var asyncWaitList = []func() error{ConsoleAsync.Close}

func Close() error {
	for i := range asyncWaitList {
		asyncWaitList[i]()
	}
	return nil
}

var stacktracePool = sync.Pool{
	New: func() interface{} {
		return newProgramCounters(64)
	},
}

type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}

var bufferPool = pool.NewBytesPool()

// TakeStacktrace is helper func to take snap short of stack trace.
func TakeStacktrace(optionalSkip ...int) string {
	skip := 0
	if len(optionalSkip) != 0 {
		skip = optionalSkip[0]
	}
	skip += depth + 2

	buff := bufferPool.Get()
	defer buff.Free()

	programCounters := stacktracePool.Get().(*programCounters)
	defer stacktracePool.Put(programCounters)

	var numFrames int
	for {
		// Skip the call to runtime.Counters and takeStacktrace so that the
		// program counters start at the caller of takeStacktrace.
		numFrames = runtime.Callers(skip, programCounters.pcs)
		if numFrames < len(programCounters.pcs) {
			break
		}
		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		programCounters = newProgramCounters(len(programCounters.pcs) * 2)
	}

	frames := runtime.CallersFrames(programCounters.pcs[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	i := 0
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if i != 0 {
			buff.AppendByte('\n')
		}
		i++
		buff.AppendString(frame.Function)
		buff.AppendByte('\n')
		buff.AppendByte('\t')

		if prefixSize != 0 && len(frame.File) > prefixSize {
			frame.File = frame.File[prefixSize:]
		}
		buff.AppendString(frame.File)

		buff.AppendByte(':')
		buff.AppendInt(int64(frame.Line))
	}

	return buff.String()
}
