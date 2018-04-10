package yspata

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
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
	LOG_DEBUG = iota
	LOG_VERBOSE
	LOG_WARN
	LOG_INFO
	LOG_SILENT
)

var logLevel = LOG_INFO

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

func Info(format string, arg ...interface{}) {
	if logLevel <= LOG_INFO {
		fmt.Printf("# %s\n", fmt.Sprintf(format, arg...))
	}
}

var onError func(error, string) = func(err error, msg string) {
	fmt.Printf("Error: %s\n", msg)
}

func SetOnError(f func(err error, msg string)) {
	onError = f
}

func callOnError(err error, format string, arg ...interface{}) {
	if f := onError; f != nil {
		f(err, fmt.Sprintf(format, arg...))
	}
}

func FailIf(err error, format string, arg ...interface{}) bool {
	if err == nil {
		return true
	} else {
		callOnError(err, format, arg...)
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
		callOnError(nil, "File '%s' is not found\n", name)
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

type CommandResult struct {
	Command string
	Args    []string
	Status  int
	Stdout  string
	Stderr  string
	Error   error
}

func newCommandResult(cmd string, args []string) *CommandResult {
	return &CommandResult{Command: cmd, Args: args}
}

func (r *CommandResult) FailIf(msg string) {
	FailIf(r.Error, msg)
}

type CommandContext struct {
	Command  string
	Args     []string
	stdin    io.WriteCloser
	stdout   io.Reader
	stderr   io.Reader
	OnStdin  func(io.WriteCloser)
	OnStdout func(io.Reader)
	OnStderr func(io.Reader)
	exec     *exec.Cmd
	result   *CommandResult
}

func (c *CommandContext) Run() *CommandResult {
	if res := c.Start(); res.Error != nil {
		return res
	}
	return c.Wait()
}

func (c *CommandContext) Start() (res *CommandResult) {
	Info("%s %s", c.Command, strings.Join(c.Args, " "))

	c.result = newCommandResult(c.Command, c.Args)
	res = c.result
	c.exec = exec.Command(c.Command, c.Args...)

	if c.OnStdin != nil {
		stdin, err := c.exec.StdinPipe()
		if err != nil {
			res.Error = err
			return
		}
		c.OnStdin(stdin)
		stdin.Close()
	}

	stdout, err := c.exec.StdoutPipe()
	res.Error = err
	if err != nil {
		return
	}
	c.stdout = stdout

	stderr, err := c.exec.StderrPipe()
	res.Error = err
	if err != nil {
		return
	}
	c.stderr = stderr

	return
}

func (c *CommandContext) Wait() (res *CommandResult) {
	res = c.result
	if err := c.exec.Start(); err != nil {
		res.Error = err
		return
	}

	var bufout, buferr bytes.Buffer
	stdout2 := io.TeeReader(c.stdout, &bufout)
	stderr2 := io.TeeReader(c.stderr, &buferr)

	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			if c.OnStdout != nil {
				c.OnStdout(stdout2)
				wg.Done()
			}
		}()
		go func() {
			if c.OnStderr != nil {
				c.OnStderr(stderr2)
				wg.Done()
			}
		}()
		wg.Wait()
	}()

	err := c.exec.Wait()
	res.Error = err
	if err != nil {
		if err2, ok := err.(*exec.ExitError); ok {
			if s, ok := err2.Sys().(syscall.WaitStatus); ok {
				res.Status = s.ExitStatus()
			}
		}
	}
	return
}

func Command(cmd string, arg ...string) *CommandContext {
	return &CommandContext{Command: cmd, Args: arg}
}

func Commandf(format string, arg ...interface{}) *CommandContext {
	s := fmt.Sprintf(format, arg...)
	comps := strings.Split(s, " ")
	return Command(comps[0], comps[1:]...)
}

func PrintOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		Printf("%s", scanner.Text())
	}
}
