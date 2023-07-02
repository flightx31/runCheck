package runcheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/flightx31/file"
	"github.com/mitchellh/go-ps"
	"github.com/spf13/afero"
	"os"
	"path"
	"strconv"
	"syscall"
)

type Logger interface {
	Panic(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
	Print(args ...interface{})
}

type L struct {
}

var l = L{}

func (l L) Panic(args ...interface{}) {
	fmt.Println("PANIC: ", args)
}

func (l L) Error(args ...interface{}) {
	fmt.Println("ERROR: ", args)
}

func (l L) Warn(args ...interface{}) {
	fmt.Println("WARN: ", args)
}

func (l L) Info(args ...interface{}) {
	fmt.Println("INFO: ", args)
}

func (l L) Debug(args ...interface{}) {
	fmt.Println("DEBUG: ", args)
}

func (l L) Trace(args ...interface{}) {
	fmt.Println("TRACE: ", args)
}

func (l L) Print(args ...interface{}) {
	fmt.Println("PRINT: ", args)
}

var log Logger = L{}

func SetLogger(l Logger) {
	log = l
}

var fs afero.Fs

func SetFs(newFs afero.Fs) {
	fs = newFs
}

// AbortStartup makes sure that only one instance of the program runs at a time
func AbortStartup(workingDirectory string, runConfigName string) (bool, error) {
	runConfigPath := path.Join(workingDirectory, runConfigName)
	file.SetFs(fs)

	if !file.ExistsAndIsFile(runConfigPath) {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}

	runningConfig, err := LoadRunningConfig(runConfigPath)
	cantLoadRunConfig := err != nil

	if cantLoadRunConfig {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}
	log.Info("Our PID: " + strconv.Itoa(os.Getpid()))
	if os.Getpid() == runningConfig.PID {
		// This is us.
		return false, nil
	}

	p, err := ps.FindProcess(runningConfig.PID)
	fmt.Print(p)

	if p == nil {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}

	//running := getProcessRunningStatus(runningConfig.PID)
	//
	//if !running {
	//	return initRunningConfigToThisPIDReturnState(runConfigPath)
	//}

	//log.Debug("Process pid " + strconv.Itoa(process.Pid) + " our pid " + strconv.Itoa(runningConfig.PID))

	return true, nil
}

// check if the process is actually running
// However, on Unix systems, os.FindProcess always succeeds and returns
// a Process for the given pid...regardless of whether the process exists
// or not.
func getProcessRunningStatus(pid int) bool {
	// Copied from: https://www.socketloop.com/tutorials/golang-get-executable-name-behind-process-id-example
	proc, err := os.FindProcess(pid)
	if err != nil {
		// On windows if the process doesn't exist, find process will return an error.
		return false
	}

	//double check if process is running and alive
	//by sending a signal 0
	//NOTE : syscall.Signal is not available in Windows

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return false
	}

	if err == syscall.ESRCH {
		return false
	}

	// default
	return true
}

func WritePortToRunConfig(port int, path string) error {
	runningConfig, err := LoadRunningConfig(path)

	if err != nil {
		return err
	}

	runningConfig.PORT = port

	err = WriteRunningConfig(path, runningConfig)

	if err != nil {
		return err
	}

	return nil
}
func initRunningConfigToThisPIDReturnState(runConfigPath string) (bool, error) {
	err := InitRunningConfigToThisPID(runConfigPath)

	if err != nil {
		log.Error("Unable to write running config")
		return true, err
	}

	return false, nil
}
func InitRunningConfigToThisPID(runConfigPath string) error {
	runningConfig := RunningConfig{}
	runningConfig.PORT = 0
	runningConfig.PID = os.Getpid()
	return WriteRunningConfig(runConfigPath, runningConfig)
}

func WritePIDToRunConfig(pid int, path string) error {
	runningConfig, err := LoadRunningConfig(path)

	if err != nil {
		runningConfig = RunningConfig{}
	}

	runningConfig.PID = pid

	err = WriteRunningConfig(path, runningConfig)

	if err != nil {
		return err
	}

	return nil
}

func DeleteRunningConfig(path string) error {
	return fs.Remove(path)
}

func LoadRunningConfig(path string) (RunningConfig, error) {
	// TODO: Find a way to read just the version number, and then switch the that versions struct to parse it in. (See this link for a solution: https://stackoverflow.com/questions/35822102/json-single-value-parsing)
	// TODO: Add a way to read just the config version number and upgrade the config - migrating or deleting fields that have been changed.
	var res = RunningConfig{}
	configBytes, err := afero.ReadFile(fs, path)
	if err != nil {
		return res, err
	}

	decoder := json.NewDecoder(bytes.NewReader(configBytes))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&res)

	if err != nil {
		return res, err
	}

	return res, nil
}

func WriteRunningConfig(path string, config RunningConfig) error {
	// See this link for pretty printing the json output: https://gosamples.dev/pretty-print-json/
	b, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	err = afero.WriteFile(fs, path, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

type RunningConfig struct {
	PID  int
	PORT int
}
