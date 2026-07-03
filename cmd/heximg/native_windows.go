//go:build windows

package main

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf8"

	"fyne.io/fyne/v2"
)

func chooseImageFile(_ fyne.Window, selected func(string, error)) {
	script := `
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '选择图片'
$dialog.Filter = '图片文件|*.jpg;*.jpeg;*.png;*.webp;*.bmp;*.tif;*.tiff;*.gif|所有文件|*.*'
$dialog.Multiselect = $false
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  Write-Output $dialog.FileName
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	hideCommandWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		selected("", errors.New("打开 Windows 文件选择器失败"))
		return
	}
	if !utf8.Valid(output) {
		selected("", errors.New("Windows 文件选择器返回了非 UTF-8 路径"))
		return
	}
	selected(strings.TrimSpace(strings.TrimPrefix(string(output), "\ufeff")), nil)
}

func hideCommandWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
