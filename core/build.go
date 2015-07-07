package core

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func GoBuild(sourceDir, goos, goarch string) (string, error) {
	buildDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	buildName := "localbar_build_" + strconv.Itoa(int(time.Now().UnixNano()))
	buildPath := filepath.Clean(buildDir + string(filepath.Separator) + buildName)

	cmd := exec.Command("go", "build", "-a", "-o", buildPath)
	cmd.Env = append(cmd.Env, "GOPATH="+os.Getenv("GOPATH"))
	cmd.Env = append(cmd.Env, "GOOS="+goos)
	cmd.Env = append(cmd.Env, "GOARCH="+goarch)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Dir = sourceDir
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return buildPath, nil
}
