package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	pathSeparator = string(os.PathSeparator)
)

func init() {
	initHost()
}

func initHost() {
	host, _ = os.Hostname()
}

func initServerPath() string {
	file, _ := exec.LookPath(os.Args[0])
	filePath, _ := filepath.Abs(file)

	index := strings.LastIndex(filePath, pathSeparator)
	deploy = filePath[:index]

	return deploy
}
