package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	defaultShellTimeout = 2 * 60 * time.Second
)

func execShell(command string) (string, error) {
	logrus.Debugf("execshell: '%s'", command)
	cmd := exec.Command("sh", "-c", command)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Error executing command '%s': '%s'", command, err)
		return "", err
	} else {
		return string(stdoutStderr), nil
	}
}

// sh is a simple os.exec Command tool, returns trimmed string output
func sh(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	logrus.Debugf("sh CMD: %q", cmd)
	out, err := cmd.Output()
	return strings.Trim(string(out), " \n"), err
}

// ShResult used for channel in timeout
type ShResult struct {
	Output string // STDOUT
	Err    error  // go error, not STDERR
}

type ShTimeoutError struct {
	timeout time.Duration
}

func (e ShTimeoutError) Error() string {
	return fmt.Sprintf("Reached TIMEOUT on shell command")
}

// shWithDefaultTimeout will use the defaultShellTimeout so you dont have to pass one
func shWithDefaultTimeout(name string, args ...string) (string, error) {
	return shWithTimeout(defaultShellTimeout, name, args...)
}

// shWithTimeout will run the Cmd and wait for the specified duration
func shWithTimeout(howLong time.Duration, name string, args ...string) (string, error) {
	// duration can't be zero
	if howLong <= 0 {
		return "", fmt.Errorf("Timeout duration needs to be positive")
	}
	// set up the results channel
	resultsChan := make(chan ShResult, 1)
	logrus.Debugf("shWithTimeout: %v, %s, %v", howLong, name, args)

	// fire up the goroutine for the actual shell command
	go func() {
		out, err := sh(name, args...)
		resultsChan <- ShResult{Output: out, Err: err}
	}()

	select {
	case res := <-resultsChan:
		return res.Output, res.Err
	case <-time.After(howLong):
		return "", ShTimeoutError{timeout: howLong}
	}

	return "", nil
}

func mkDirs(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func joinLastLines(lines []string, n int) string {
	return strings.Join(lines[len(lines)-1-n:], "\n")
}
