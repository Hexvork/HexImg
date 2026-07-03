//go:build windows

package main

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"
)

func chooseImageFile() (string, error) {
	script := `
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '选择图片'
$dialog.Filter = '图片文件|*.jpg;*.jpeg;*.png;*.webp;*.bmp;*.tif;*.tiff;*.gif|所有文件|*.*'
$dialog.Multiselect = $false
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
  Write-Output $dialog.FileName
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	hideCommandWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("打开 Windows 文件选择器失败")
	}
	return strings.TrimSpace(string(output)), nil
}

func hideCommandWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
