package yspata

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var IsMac = isMac()

var IsLinux = isLinux()

func isMac() bool {
	return runtime.GOOS == "darwin"
}

func isLinux() bool {
	return runtime.GOOS == "linux"
}

func FullVersion(version string) string {
	return fmt.Sprintf("%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
}

func Success() {
	os.Exit(0)
}

func Fail() {
	os.Exit(1)
}

func Printf(format string, arg ...interface{}) {
	fmt.Printf("%s\n", fmt.Sprintf(format, arg...))
}

func Eprintf(format string, arg ...interface{}) {
	fmt.Printf("Error: %s\n", fmt.Sprintf(format, arg...))
}

func PrintLines(arg ...string) {
	for _, s := range arg {
		fmt.Println(s)
	}
}

const (
	LOG_DEBUG   = 0
	LOG_VERBOSE = 1
	LOG_WARN    = 2
	LOG_SILENT  = 3
)

var logLevel = LOG_WARN

func SetLogLevel(lv int) {
	logLevel = lv
}

func Debug(format string, arg ...interface{}) {
	if logLevel <= LOG_DEBUG {
		fmt.Printf("# %s\n", fmt.Sprintf(format, arg...))
	}
}

func Verbose(format string, arg ...interface{}) {
	if logLevel <= LOG_VERBOSE {
		fmt.Printf("# %s\n", fmt.Sprintf(format, arg...))
	}
}

func Warn(format string, arg ...interface{}) {
	if logLevel <= LOG_WARN {
		fmt.Printf("# %s\n", fmt.Sprintf(format, arg...))
	}
}

type Result struct {
	Error error
}

func Eval(err error) *Result {
	return &Result{Error: err}
}

func Eval2(_ interface{}, err error) *Result {
	return &Result{Error: err}
}

func FailIf(err error, format string, arg ...interface{}) bool {
	if err == nil {
		return true
	} else {
		fmt.Printf("Error: %s\n", fmt.Sprintf(format, arg...))
		Fail()
		return false
	}
}

func Join(elem ...string) string {
	return filepath.Join(elem...)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func FailIfNotExists(name string) {
	if !Exists(name) {
		fmt.Printf("Error: File '%s' is not found\n", name)
		os.Exit(1)
	}
}

func Open(name string) *os.File {
	return OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
}

func OpenFile(name string, flags int, perm os.FileMode) *os.File {
	FailIfNotExists(name)
	file, err := os.OpenFile(name, flags, perm)
	FailIf(err, err.Error())
	return file
}

func Contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
