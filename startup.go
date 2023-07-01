package runcheck

import (
	"bytes"
	"encoding/json"
	"github.com/flightx31/exception"
	"github.com/flightx31/file"
	"github.com/spf13/afero"
	"os"
	"path"
	"strconv"
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

var log Logger

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

	if !file.FileExists(runConfigPath) {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}

	runningConfig, err := LoadRunningConfig(runConfigPath)
	cantLoadRunConfig := err != nil

	if cantLoadRunConfig {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}

	// Always returns a process on linux/OSX, on Windows it fails if that process doesn't exist.
	process, err := os.FindProcess(runningConfig.PID)
	exception.LogError(err)
	log.Debug("Process pid " + strconv.Itoa(process.Pid) + " our pid " + strconv.Itoa(runningConfig.PID))

	if err != nil {
		return initRunningConfigToThisPIDReturnState(runConfigPath)
	}
	return false, nil
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
