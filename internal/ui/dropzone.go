package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// DropZone is a widget with a dashed border for drag-and-drop
type DropZone struct {
	widget.BaseWidget
	content     fyne.CanvasObject
	borderColor color.Color
	bgColor     color.Color
}

// NewDropZone creates a new drop zone widget
func NewDropZone(content fyne.CanvasObject, dark bool) *DropZone {
	dz := &DropZone{
		content: content,
	}
	if dark {
		dz.borderColor = color.NRGBA{R: 0x8B, G: 0x94, B: 0x9E, A: 0xAA}
		dz.bgColor = color.NRGBA{R: 0x16, G: 0x1B, B: 0x22, A: 0xFF}
	} else {
		dz.borderColor = color.NRGBA{R: 0x6B, G: 0x72, B: 0x80, A: 0xAA}
		dz.bgColor = color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF}
	}
	dz.ExtendBaseWidget(dz)
	return dz
}

func (dz *DropZone) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(dz.bgColor)
	bg.CornerRadius = 8

	return &dropZoneRenderer{
		dropZone: dz,
		bg:       bg,
		lines:    make([]*canvas.Line, 0),
		objects:  []fyne.CanvasObject{bg, dz.content},
	}
}

type dropZoneRenderer struct {
	dropZone *DropZone
	bg       *canvas.Rectangle
	lines    []*canvas.Line
	objects  []fyne.CanvasObject
}

func (r *dropZoneRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)

	padding := float32(12)
	r.dropZone.content.Move(fyne.NewPos(padding, padding))
	r.dropZone.content.Resize(fyne.NewSize(size.Width-2*padding, size.Height-2*padding))

	// Clear old lines
	r.lines = r.lines[:0]
	r.objects = []fyne.CanvasObject{r.bg, r.dropZone.content}

	// Create dashed border
	dashLength := float32(8)
	gapLength := float32(6)
	strokeWidth := float32(1.5)
	offset := strokeWidth / 2

	// Top edge
	x := offset
	for x < size.Width-offset {
		lineEnd := x + dashLength
		if lineEnd > size.Width-offset {
			lineEnd = size.Width - offset
		}
		line := canvas.NewLine(r.dropZone.borderColor)
		line.StrokeWidth = strokeWidth
		line.Position1 = fyne.NewPos(x, offset)
		line.Position2 = fyne.NewPos(lineEnd, offset)
		r.lines = append(r.lines, line)
		r.objects = append(r.objects, line)
		x = lineEnd + gapLength
	}

	// Bottom edge
	x = offset
	for x < size.Width-offset {
		lineEnd := x + dashLength
		if lineEnd > size.Width-offset {
			lineEnd = size.Width - offset
		}
		line := canvas.NewLine(r.dropZone.borderColor)
		line.StrokeWidth = strokeWidth
		line.Position1 = fyne.NewPos(x, size.Height-offset)
		line.Position2 = fyne.NewPos(lineEnd, size.Height-offset)
		r.lines = append(r.lines, line)
		r.objects = append(r.objects, line)
		x = lineEnd + gapLength
	}

	// Left edge
	y := offset
	for y < size.Height-offset {
		lineEnd := y + dashLength
		if lineEnd > size.Height-offset {
			lineEnd = size.Height - offset
		}
		line := canvas.NewLine(r.dropZone.borderColor)
		line.StrokeWidth = strokeWidth
		line.Position1 = fyne.NewPos(offset, y)
		line.Position2 = fyne.NewPos(offset, lineEnd)
		r.lines = append(r.lines, line)
		r.objects = append(r.objects, line)
		y = lineEnd + gapLength
	}

	// Right edge
	y = offset
	for y < size.Height-offset {
		lineEnd := y + dashLength
		if lineEnd > size.Height-offset {
			lineEnd = size.Height - offset
		}
		line := canvas.NewLine(r.dropZone.borderColor)
		line.StrokeWidth = strokeWidth
		line.Position1 = fyne.NewPos(size.Width-offset, y)
		line.Position2 = fyne.NewPos(size.Width-offset, lineEnd)
		r.lines = append(r.lines, line)
		r.objects = append(r.objects, line)
		y = lineEnd + gapLength
	}
}

func (r *dropZoneRenderer) MinSize() fyne.Size {
	contentMin := r.dropZone.content.MinSize()
	padding := float32(24)
	return fyne.NewSize(contentMin.Width+padding, contentMin.Height+padding)
}

func (r *dropZoneRenderer) Refresh() {
	r.bg.FillColor = r.dropZone.bgColor
	r.bg.Refresh()
	canvas.Refresh(r.dropZone)
}

func (r *dropZoneRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *dropZoneRenderer) Destroy() {}
