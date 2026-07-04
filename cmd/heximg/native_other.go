//go:build !windows

package main

import (
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

func chooseImageFile(win fyne.Window, selected func([]string, error)) {
	chooseImageFileInDirectory(win, "", selected)
}

func chooseImageFileInDirectory(win fyne.Window, dir string, selected func([]string, error)) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			selected(nil, err)
			return
		}
		if reader == nil {
			selected(nil, nil)
			return
		}
		defer reader.Close()
		selected([]string{reader.URI().Path()}, nil)
	}, win)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{
		".jpg", ".jpeg", ".png", ".webp", ".avif", ".gif", ".bmp", ".tif", ".tiff", ".ico", ".svg",
	}))
	if dir != "" {
		if listable, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
			fileDialog.SetLocation(listable)
		}
	}
	fileDialog.Show()
}

func hideCommandWindow(cmd *exec.Cmd) {
}
