//go:build windows

package main

import (
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
)

func chooseImageFile(win fyne.Window, selected func([]string, error)) {
	chooseImageFileInDirectory(win, "", selected)
}

func chooseImageFileInDirectory(win fyne.Window, dir string, selected func([]string, error)) {
	script := `
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.OpenFileDialog
$dialog.Title = '选择图片'
$dialog.Filter = '图片文件|*.jpg;*.jpeg;*.png;*.webp;*.bmp;*.tif;*.tiff;*.gif|所有文件|*.*'
$dialog.Multiselect = $true
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
  Write-Output ($dialog.FileNames -join '|')
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	hideCommandWindow(cmd)
	output, err := cmd.Output()
	if err != nil {
		selected(nil, nil)
		return
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		selected(nil, nil)
		return
	}
	paths := strings.Split(text, "|")
	var filtered []string
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		selected(nil, nil)
		return
	}
	selected(filtered, nil)
}

func hideCommandWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
