//go:build windows

package main

import (
	"os/exec"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func chooseImageFile(win fyne.Window, selected func(string, error)) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			selected("", err)
			return
		}
		if reader == nil {
			selected("", nil)
			return
		}
		defer reader.Close()
		selected(reader.URI().Path(), nil)
	}, win)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{
		".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tif", ".tiff", ".gif",
	}))
	fileDialog.Show()
}

func hideCommandWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
