//go:build !windows

package main

import (
	"errors"
	"os/exec"
)

func chooseImageFile() (string, error) {
	return "", errors.New("当前原生文件选择器仅支持 Windows")
}

func hideCommandWindow(cmd *exec.Cmd) {
}
