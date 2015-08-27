package log

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// severity identifies the sort of log: info, warning, etc. It also implements
// the flag.Value interface. The -v flag is of type severity and should be
// modified only through the flag.Value interface.
type severity int32 // sync/atomic int32

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log file is also written to each
// lower-severity log file.
const (
	traceLog severity = iota
	debugLog
	infoLog
	warningLog
	errorLog
	fatalLog
	numSeverity = 5
)

const severityChar = "TDIWEF"

var severityName = []string{
	traceLog:   "TRACE",
	debugLog:   "DEBUG",
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

// get returns the value of the severity.
func (s *severity) get() severity {
	return severity(atomic.LoadInt32((*int32)(s)))
}

// set sets the value of the severity.
func (s *severity) set(val severity) {
	atomic.StoreInt32((*int32)(s), int32(val))
}

// String is part of the flag.Value interface.
func (s *severity) String() string {
	return strconv.FormatInt(int64(*s), 10)
}

// Get is part of the flag.Value interface.
func (s *severity) Get() interface{} {
	return *s
}

var errSeverity = fmt.Errorf("valid values are: %v", severityName)

// Set is part of the flag.Value interface.
func (s *severity) Set(value string) error {
	value = strings.ToUpper(value)
	v := -1
	for i, name := range severityName {
		if name == value {
			v = i
			break
		}
	}
	if v == -1 {
		return errSeverity
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	logging.verbosity.set(severity(v))
	return nil
}

func severityByName(s string) (severity, bool) {
	s = strings.ToUpper(s)
	for i, name := range severityName {
		if name == s {
			return severity(i), true
		}
	}
	return 0, false
}

// OutputStats tracks the number of output lines and bytes written.
type OutputStats struct {
	lines int64
	bytes int64
}

// Lines returns the number of lines written.
func (s *OutputStats) Lines() int64 {
	return atomic.LoadInt64(&s.lines)
}

// Bytes returns the number of bytes written.
func (s *OutputStats) Bytes() int64 {
	return atomic.LoadInt64(&s.bytes)
}

// Stats tracks the number of lines of output and number of bytes
// per severity level. Values must be read with atomic.LoadInt64.
var Stats struct {
	Trace, Debug, Info, Warning, Error OutputStats
}

var severityStats = [numSeverity]*OutputStats{
	traceLog:   &Stats.Trace,
	debugLog:   &Stats.Debug,
	infoLog:    &Stats.Info,
	warningLog: &Stats.Warning,
	errorLog:   &Stats.Error,
}

// traceLocation represents the setting of the -log_backtrace_at flag.
type traceLocation struct {
	file string
	line int
}

// isSet reports whether the trace location has been specified.
// logging.mu is held.
func (t *traceLocation) isSet() bool {
	return t.line > 0
}

// match reports whether the specified file and line matches the trace location.
// The argument file name is the full path, not the basename specified in the flag.
// logging.mu is held.
func (t *traceLocation) match(file string, line int) bool {
	if t.line != line {
		return false
	}
	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}
	return t.file == file
}

func (t *traceLocation) String() string {
	// Lock because the type is not atomic. TODO: clean this up.
	logging.mu.Lock()
	defer logging.mu.Unlock()
	return fmt.Sprintf("%s:%d", t.file, t.line)
}

// Get is part of the (Go 1.2) flag.Getter interface. It always returns nil for this flag type since the
// struct is not exported
func (t *traceLocation) Get() interface{} {
	return nil
}

var errTraceSyntax = errors.New("syntax error: expect file.go:234")

// Syntax: -log_backtrace_at=gopherflakes.go:234
// Note the file extension is included here.
func (t *traceLocation) Set(value string) error {
	if value == "" {
		// Unset.
		t.line = 0
		t.file = ""
	}
	fields := strings.Split(value, ":")
	if len(fields) != 2 {
		return errTraceSyntax
	}
	file, line := fields[0], fields[1]
	if !strings.Contains(file, ".") {
		return errTraceSyntax
	}
	v, err := strconv.Atoi(line)
	if err != nil {
		return errTraceSyntax
	}
	if v <= 0 {
		return errors.New("negative or zero value for level")
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	t.line = v
	t.file = file
	return nil
}

func init() {
	logging.verbosity = infoLog
	flag.Var(&logging.verbosity, "v", "log level")
	flag.Var(&logging.traceLocation, "log_backtrace_at", "when logging hits line file:N, emit a stack trace")
}

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	// Boolean flags. Not handled atomically because the flag.Value interface
	// does not let us avoid the =true, and that shorthand is necessary for
	// compatibility. TODO: does this matter enough to fix? Seems unlikely.

	// freeList is a list of byte buffers, maintained under freeListMu.
	freeList *buffer
	// freeListMu maintains the free list. It is separate from the main mutex
	// so buffers can be grabbed and printed to without holding the main lock,
	// for better parallelization.
	freeListMu sync.Mutex

	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// pcs is used in V to avoid an allocation when computing the caller's PC.
	pcs [1]uintptr
	// traceLocation is the state of the -log_backtrace_at flag.
	traceLocation traceLocation
	// These flags are modified only under lock, although verbosity may be fetched
	// safely using atomic.LoadInt32.
	verbosity severity // logging level, the value of the -v flag
}

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	next *buffer
}

var logging loggingT

// getBuffer returns a new, ready-to-use buffer.
func (l *loggingT) getBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (l *loggingT) putBuffer(b *buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}
	l.freeListMu.Lock()
	b.next = l.freeList
	l.freeList = b
	l.freeListMu.Unlock()
}

var timeNow = time.Now // Stubbed out for testing.

func (l *loggingT) header(s severity, depth int) (*buffer, string, int) {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return l.formatHeader(s, file, line), file, line
}

var pid = os.Getpid()

// formatHeader formats a log header using the provided file name and line number.
func (l *loggingT) formatHeader(s severity, file string, line int) *buffer {
	if s > fatalLog {
		s = infoLog // for safety.
	}
	buf := l.getBuffer()
	// Lfile:line]
	buf.WriteString(string(severityChar[s]) + " " + file + ":" + strconv.Itoa(line) + "] ")
	return buf
}

// printX funcs are named pX because go vet is not very smart and complains
// about s not being a string

func (l *loggingT) pln(s severity, args ...interface{}) {
	buf, file, line := l.header(s, 0)
	fmt.Fprintln(buf, args...)
	l.output(s, buf, file, line)
}

func (l *loggingT) p(s severity, args ...interface{}) {
	l.pDepth(s, 1, args...)
}

func (l *loggingT) pDepth(s severity, depth int, args ...interface{}) {
	buf, file, line := l.header(s, depth)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, file, line)
}

func (l *loggingT) pf(s severity, format string, args ...interface{}) {
	buf, file, line := l.header(s, 0)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, file, line)
}

// pWithFileLine behaves like print but uses the provided file and line number.
func (l *loggingT) pWithFileLine(s severity, file string, line int, args ...interface{}) {
	buf := l.formatHeader(s, file, line)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, file, line)
}

// output writes the data to the log files and releases the buffer.
func (l *loggingT) output(s severity, buf *buffer, file string, line int) {
	l.mu.Lock()
	if l.traceLocation.isSet() {
		if l.traceLocation.match(file, line) {
			buf.Write(stacks(false))
		}
	}
	data := buf.Bytes()
	os.Stderr.Write(data)
	if s == fatalLog {
		os.Stderr.Write(stacks(true))
		os.Exit(255)
	}
	l.putBuffer(buf)
	l.mu.Unlock()
	if stats := severityStats[s]; stats != nil {
		atomic.AddInt64(&stats.lines, 1)
		atomic.AddInt64(&stats.bytes, int64(len(data)))
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't fit. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// logExitFunc provides a simple mechanism to override the default behavior
// of exiting on error. Used in testing and to guarantee we reach a required exit
// for fatal logs. Instead, exit could be a function rather than a method but that
// would make its use clumsier.
var logExitFunc func(error)

func Trace(args ...interface{}) {
	if traceLog >= logging.verbosity {
		logging.p(traceLog, args...)
	}
}

func TraceDepth(depth int, args ...interface{}) {
	if traceLog >= logging.verbosity {
		logging.pDepth(traceLog, depth, args...)
	}
}

func Traceln(args ...interface{}) {
	if traceLog >= logging.verbosity {
		logging.pln(traceLog, args...)
	}
}

func Tracef(format string, args ...interface{}) {
	if traceLog >= logging.verbosity {
		logging.pf(traceLog, format, args...)
	}
}

func Debug(args ...interface{}) {
	if debugLog >= logging.verbosity {
		logging.p(debugLog, args...)
	}
}

func DebugDepth(depth int, args ...interface{}) {
	if debugLog >= logging.verbosity {
		logging.pDepth(debugLog, depth, args...)
	}
}

func Debugln(args ...interface{}) {
	if debugLog >= logging.verbosity {
		logging.pln(debugLog, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if debugLog >= logging.verbosity {
		logging.pf(debugLog, format, args...)
	}
}

func Info(args ...interface{}) {
	if infoLog >= logging.verbosity {
		logging.p(infoLog, args...)
	}
}

func InfoDepth(depth int, args ...interface{}) {
	if infoLog >= logging.verbosity {
		logging.pDepth(infoLog, depth, args...)
	}
}

func Infoln(args ...interface{}) {
	if infoLog >= logging.verbosity {
		logging.pln(infoLog, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if infoLog >= logging.verbosity {
		logging.pf(infoLog, format, args...)
	}
}

func Warning(args ...interface{}) {
	if warningLog >= logging.verbosity {
		logging.p(warningLog, args...)
	}
}

func WarningDepth(depth int, args ...interface{}) {
	if warningLog >= logging.verbosity {
		logging.pDepth(warningLog, depth, args...)
	}
}

func Warningln(args ...interface{}) {
	if warningLog >= logging.verbosity {
		logging.pln(warningLog, args...)
	}
}

func Warningf(format string, args ...interface{}) {
	if warningLog >= logging.verbosity {
		logging.pf(warningLog, format, args...)
	}
}

func Error(args ...interface{}) {
	if errorLog >= logging.verbosity {
		logging.p(errorLog, args...)
	}
}

func ErrorDepth(depth int, args ...interface{}) {
	if errorLog >= logging.verbosity {
		logging.pDepth(errorLog, depth, args...)
	}
}

func Errorln(args ...interface{}) {
	if errorLog >= logging.verbosity {
		logging.pln(errorLog, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if errorLog >= logging.verbosity {
		logging.pf(errorLog, format, args...)
	}
}

func Fatal(args ...interface{}) {
	if fatalLog >= logging.verbosity {
		logging.p(fatalLog, args...)
	}
}

func FatalDepth(depth int, args ...interface{}) {
	if fatalLog >= logging.verbosity {
		logging.pDepth(fatalLog, depth, args...)
	}
}

func Fatalln(args ...interface{}) {
	if fatalLog >= logging.verbosity {
		logging.pln(fatalLog, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if fatalLog >= logging.verbosity {
		logging.pf(fatalLog, format, args...)
	}
}

func Atrace(ln string) {
	logging.p(traceLog, ln)
}

func Version(revision string) {
	ln := fmt.Sprintf("%s %s %s", os.Args[0], revision, runtime.Version())
	logging.p(infoLog, ln)
}
