package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fas "github.com/Hexvork/HexImg/internal/fas"
	"github.com/Hexvork/HexImg/internal/ui"
)

var version = "dev"

type convertConfig struct {
	input   string
	output  string
	format  string
	quality int
}

func main() {
	hex := app.NewWithID("com.hexvork.heximg")
	darkMode := true
	hex.Settings().SetTheme(ui.NewFluentTheme(darkMode))

	win := hex.NewWindow("HexImg")
	win.SetIcon(fas.Image)
	win.Resize(fyne.NewSize(760, 520))
	win.SetMaster()

	selectedPath := widget.NewEntry()
	selectedPath.SetPlaceHolder("请选择一张图片")
	selectedPath.Disable()

	outputPath := widget.NewEntry()
	outputPath.SetPlaceHolder("输出路径会自动生成")
	outputPath.Disable()

	statusLabel := widget.NewLabel("就绪")
	statusLabel.Wrapping = fyne.TextWrapWord

	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.SetMinRowsVisible(5)
	logOutput.Disable()

	formatSelect := widget.NewSelect([]string{"jpg", "png", "webp", "bmp", "tiff"}, nil)
	formatSelect.SetSelected("jpg")

	qualityLabel := widget.NewLabel("质量")
	qualityValue := widget.NewLabel("85")
	qualitySlider := widget.NewSlider(0, 100)
	qualitySlider.Step = 1
	qualitySlider.Value = 85
	qualitySlider.OnChanged = func(value float64) {
		qualityValue.SetText(strconv.Itoa(int(value)))
	}

	cfg := func() convertConfig {
		input := strings.TrimSpace(selectedPath.Text)
		format := formatSelect.Selected
		return convertConfig{
			input:   input,
			output:  outputFor(input, format),
			format:  format,
			quality: int(qualitySlider.Value),
		}
	}

	refreshOutput := func() {
		outputPath.SetText(cfg().output)
	}

	lastFormat := formatSelect.Selected
	lastLossyQuality := 85.0
	lastPngCompression := 6.0
	refreshQualityControl := func(format string) {
		if format == "png" {
			qualityLabel.SetText("压缩级别")
			qualitySlider.Min = 0
			qualitySlider.Max = 9
			qualitySlider.Step = 1
			qualitySlider.Value = lastPngCompression
		} else {
			qualityLabel.SetText("质量")
			qualitySlider.Min = 0
			qualitySlider.Max = 100
			qualitySlider.Step = 1
			qualitySlider.Value = lastLossyQuality
		}
		qualityValue.SetText(strconv.Itoa(int(qualitySlider.Value)))
		qualityLabel.Refresh()
		qualitySlider.Refresh()
		qualityValue.Refresh()
	}

	formatSelect.OnChanged = func(format string) {
		if lastFormat == "png" {
			lastPngCompression = qualitySlider.Value
		} else {
			lastLossyQuality = qualitySlider.Value
		}
		refreshQualityControl(format)
		lastFormat = format
		refreshOutput()
	}

	openInputButton := fixedButton("选择图片", icon(fas.FolderOpen), func() {
		path, err := chooseImageFile()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		if path == "" {
			return
		}
		selectedPath.SetText(path)
		refreshOutput()
		statusLabel.SetText("已选择：" + filepath.Base(path))
	})

	var cancelMu sync.Mutex
	var cancelRun context.CancelFunc

	convertButton := widget.NewButtonWithIcon("转换", icon(fas.Play), nil)
	convertButton.Importance = widget.HighImportance
	convertButtonContainer := fixedButtonObject(convertButton)

	cancelButton := widget.NewButtonWithIcon("停止", icon(fas.Stop), func() {
		cancelMu.Lock()
		defer cancelMu.Unlock()
		if cancelRun != nil {
			cancelRun()
			statusLabel.SetText("正在停止转换...")
		}
	})
	cancelButton.Disable()
	cancelButtonContainer := fixedButtonObject(cancelButton)

	convertButton.OnTapped = func() {
		current := cfg()
		if current.input == "" {
			dialog.ShowError(errors.New("请先选择图片"), win)
			return
		}
		if current.format == "" {
			dialog.ShowError(errors.New("请先选择转换格式"), win)
			return
		}
		if _, err := os.Stat(current.input); err != nil {
			dialog.ShowError(fmt.Errorf("输入图片不可用：%w", err), win)
			return
		}

		args := buildFFmpegArgs(current)
		logOutput.SetText("")
		statusLabel.SetText("正在转换...")
		convertButton.Disable()
		cancelButton.Enable()

		ctx, cancel := context.WithCancel(context.Background())
		cancelMu.Lock()
		cancelRun = cancel
		cancelMu.Unlock()

		go func() {
			err := runFFmpeg(ctx, args, func(line string) {
				appendLog(logOutput, line)
			})

			cancelMu.Lock()
			cancelRun = nil
			cancelMu.Unlock()

			fyne.Do(func() {
				convertButton.Enable()
				cancelButton.Disable()
			})

			if errors.Is(ctx.Err(), context.Canceled) {
				setStatus(statusLabel, "已停止")
				return
			}
			if err != nil {
				setStatus(statusLabel, "转换失败："+err.Error())
				return
			}
			setStatus(statusLabel, "转换完成："+filepath.Base(current.output))
		}()
	}

	title := canvas.NewText("HexImg", ui.TextColor(darkMode))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText(fmt.Sprintf("图片格式转换 · FFmpeg · %s · %s/%s", version, runtime.GOOS, runtime.GOARCH), ui.MutedTextColor(darkMode))
	subtitle.TextSize = 13

	var refreshCustomTheme func()
	var themeButton *widget.Button
	themeButton = widget.NewButtonWithIcon("", icon(fas.Sun), func() {
		darkMode = !darkMode
		hex.Settings().SetTheme(ui.NewFluentTheme(darkMode))
		if darkMode {
			themeButton.Icon = icon(fas.Sun)
		} else {
			themeButton.Icon = icon(fas.Moon)
		}
		themeButton.Refresh()
		refreshCustomTheme()
	})

	header := container.NewBorder(nil, nil, nil, fixedIconButton(themeButton), container.NewVBox(title, subtitle))
	sourceCard, refreshSourceCard := fluentCard("图片", container.NewVBox(
		widget.NewLabel("输入图片"),
		container.NewBorder(nil, nil, nil, openInputButton, selectedPath),
		widget.NewLabel("输出文件"),
		outputPath,
	), darkMode)

	settingsCard, refreshSettingsCard := fluentCard("转换设置", container.NewVBox(
		widget.NewLabel("目标格式"),
		formatSelect,
		container.NewBorder(nil, nil, qualityLabel, qualityValue, qualitySlider),
	), darkMode)

	logCard, refreshLogCard := fluentCard("状态", container.NewVBox(statusLabel, logOutput), darkMode)
	actionBar := container.NewBorder(nil, nil, nil, container.NewHBox(cancelButtonContainer, convertButtonContainer), nil)

	refreshCustomTheme = func() {
		title.Color = ui.TextColor(darkMode)
		subtitle.Color = ui.MutedTextColor(darkMode)
		title.Refresh()
		subtitle.Refresh()
		refreshSourceCard(darkMode)
		refreshSettingsCard(darkMode)
		refreshLogCard(darkMode)
	}

	content := container.NewBorder(
		header,
		actionBar,
		nil,
		nil,
		container.NewPadded(container.NewVBox(sourceCard, settingsCard, logCard)),
	)
	win.SetContent(content)
	refreshOutput()
	refreshQualityControl(formatSelect.Selected)
	win.ShowAndRun()
}

func fixedButton(text string, iconResource fyne.Resource, tapped func()) fyne.CanvasObject {
	return fixedButtonObject(widget.NewButtonWithIcon(text, iconResource, tapped))
}

func fixedButtonObject(button *widget.Button) fyne.CanvasObject {
	return container.NewGridWrap(fyne.NewSize(128, 40), button)
}

func fixedIconButton(button *widget.Button) fyne.CanvasObject {
	return container.NewGridWrap(fyne.NewSize(40, 40), button)
}

func fluentCard(titleText string, content fyne.CanvasObject, darkMode bool) (fyne.CanvasObject, func(bool)) {
	title := canvas.NewText(titleText, ui.TextColor(darkMode))
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 18

	body := container.NewVBox(title, widget.NewSeparator(), content)
	bg := canvas.NewRectangle(ui.CardColor(darkMode))
	bg.CornerRadius = 8
	card := container.NewStack(bg, container.NewPadded(body))
	refresh := func(dark bool) {
		title.Color = ui.TextColor(dark)
		bg.FillColor = ui.CardColor(dark)
		title.Refresh()
		bg.Refresh()
	}
	return card, refresh
}

func icon(resource fyne.Resource) fyne.Resource {
	return theme.NewThemedResource(resource)
}

func outputFor(inputPath, format string) string {
	if inputPath == "" {
		return ""
	}
	if format == "" {
		format = "jpg"
	}
	base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	return base + "_converted." + format
}

func buildFFmpegArgs(cfg convertConfig) []string {
	args := []string{"-hide_banner", "-y", "-i", cfg.input}

	switch cfg.format {
	case "jpg", "jpeg":
		args = append(args, "-frames:v", "1", "-q:v", strconv.Itoa(jpegQScale(cfg.quality)))
	case "webp":
		args = append(args, "-frames:v", "1", "-compression_level", "6", "-quality", strconv.Itoa(cfg.quality))
	case "png":
		args = append(args, "-frames:v", "1", "-compression_level", strconv.Itoa(pngCompressionLevel(cfg.quality)))
	case "bmp", "tiff":
		args = append(args, "-frames:v", "1")
	default:
		args = append(args, "-frames:v", "1")
	}

	return append(args, cfg.output)
}

func jpegQScale(quality int) int {
	quality = clampQuality(quality)
	return 31 - int(float64(quality)*29.0/100.0)
}

func pngCompressionLevel(value int) int {
	if value < 0 {
		return 0
	}
	if value > 9 {
		return 9
	}
	return value
}

func clampQuality(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func runFFmpeg(ctx context.Context, args []string, appendLine func(string)) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	hideCommandWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("未找到 ffmpeg，请先安装 FFmpeg 并加入 PATH")
		}
		return err
	}

	var wg sync.WaitGroup
	readPipe := func(reader io.Reader) error {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			appendLine(scanner.Text())
		}
		return scanner.Err()
	}
	wg.Add(2)
	errs := make(chan error, 2)
	go func() { errs <- readPipe(stdout) }()
	go func() { errs <- readPipe(stderr) }()

	waitErr := cmd.Wait()
	wg.Wait()
	close(errs)
	for scanErr := range errs {
		if scanErr != nil && waitErr == nil {
			return scanErr
		}
	}
	return waitErr
}

func appendLog(logOutput *widget.Entry, line string) {
	fyne.Do(func() {
		const maxLogLength = 24000
		current := logOutput.Text
		runes := []rune(current)
		if len(runes) > maxLogLength {
			current = string(runes[len(runes)-maxLogLength:])
		}
		logOutput.SetText(current + line + "\n")
	})
}

func setStatus(label *widget.Label, text string) {
	fyne.Do(func() {
		label.SetText(text)
	})
}
