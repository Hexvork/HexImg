package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
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

type encodeConfig struct {
	input     string
	output    string
	format    string
	quality   string
	preset    string
	scale     string
	extraArgs string
	overwrite bool
}

func main() {
	hex := app.NewWithID("com.hexvork.heximg")
	darkMode := true
	hex.Settings().SetTheme(ui.NewFluentTheme(darkMode))

	win := hex.NewWindow("HexImg")
	win.SetIcon(fas.Image)
	win.Resize(fyne.NewSize(1120, 720))
	win.SetMaster()

	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("选择图片、视频或音频文件")

	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("选择输出文件")

	formatSelect := widget.NewSelect([]string{"mp4", "webm", "mkv", "gif", "mp3", "wav"}, nil)
	formatSelect.SetSelected("mp4")

	qualitySelect := widget.NewSelect([]string{"高质量", "平衡", "体积优先"}, nil)
	qualitySelect.SetSelected("平衡")

	presetSelect := widget.NewSelect([]string{"auto", "ultrafast", "veryfast", "faster", "fast", "medium", "slow"}, nil)
	presetSelect.SetSelected("medium")

	scaleEntry := widget.NewEntry()
	scaleEntry.SetPlaceHolder("例如 1920:-1，留空则保持原尺寸")

	extraArgsEntry := widget.NewEntry()
	extraArgsEntry.SetPlaceHolder("额外参数，例如 -movflags +faststart")

	overwriteCheck := widget.NewCheck("覆盖已存在文件", nil)
	overwriteCheck.SetChecked(true)

	commandPreview := widget.NewMultiLineEntry()
	commandPreview.Wrapping = fyne.TextWrapWord
	commandPreview.SetMinRowsVisible(4)
	commandPreview.Disable()

	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.SetMinRowsVisible(10)
	logOutput.Disable()

	statusLabel := widget.NewLabel("就绪")
	statusLabel.Wrapping = fyne.TextWrapWord

	var themeButton *widget.Button
	themeButton = widget.NewButtonWithIcon("浅色模式", icon(fas.Sun), func() {
		darkMode = !darkMode
		hex.Settings().SetTheme(ui.NewFluentTheme(darkMode))
		if darkMode {
			themeButton.Text = "浅色模式"
			themeButton.Icon = icon(fas.Sun)
			themeButton.Refresh()
			return
		}
		themeButton.Text = "深色模式"
		themeButton.Icon = icon(fas.Moon)
		themeButton.Refresh()
	})

	cfg := func() encodeConfig {
		return encodeConfig{
			input:     strings.TrimSpace(inputEntry.Text),
			output:    strings.TrimSpace(outputEntry.Text),
			format:    formatSelect.Selected,
			quality:   qualitySelect.Selected,
			preset:    presetSelect.Selected,
			scale:     strings.TrimSpace(scaleEntry.Text),
			extraArgs: strings.TrimSpace(extraArgsEntry.Text),
			overwrite: overwriteCheck.Checked,
		}
	}

	refreshPreview := func(_ string) {
		args := buildFFmpegArgs(cfg())
		commandPreview.SetText(shellPreview("ffmpeg", args))
	}

	inputEntry.OnChanged = func(value string) {
		if strings.TrimSpace(outputEntry.Text) == "" && strings.TrimSpace(value) != "" {
			outputEntry.SetText(defaultOutputPath(value, formatSelect.Selected))
			return
		}
		refreshPreview(value)
	}
	outputEntry.OnChanged = refreshPreview
	scaleEntry.OnChanged = refreshPreview
	extraArgsEntry.OnChanged = refreshPreview
	formatSelect.OnChanged = func(value string) {
		if strings.TrimSpace(inputEntry.Text) != "" {
			outputEntry.SetText(defaultOutputPath(inputEntry.Text, value))
			return
		}
		refreshPreview(value)
	}
	qualitySelect.OnChanged = refreshPreview
	presetSelect.OnChanged = refreshPreview
	overwriteCheck.OnChanged = func(bool) { refreshPreview("") }
	refreshPreview("")

	openInputButton := widget.NewButtonWithIcon("选择输入", icon(fas.FolderOpen), func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			inputEntry.SetText(reader.URI().Path())
		}, win).Show()
	})

	openOutputButton := widget.NewButtonWithIcon("保存为", icon(fas.FileExport), func() {
		dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			outputEntry.SetText(writer.URI().Path())
		}, win).Show()
	})

	var cancelMu sync.Mutex
	var cancelRun context.CancelFunc

	runButton := widget.NewButtonWithIcon("开始处理", icon(fas.Play), nil)
	runButton.Importance = widget.HighImportance

	cancelButton := widget.NewButtonWithIcon("停止", icon(fas.Stop), func() {
		cancelMu.Lock()
		defer cancelMu.Unlock()
		if cancelRun != nil {
			cancelRun()
			statusLabel.SetText("正在停止 FFmpeg...")
		}
	})
	cancelButton.Disable()

	runButton.OnTapped = func() {
		current := cfg()
		if current.input == "" {
			dialog.ShowError(errors.New("请先选择输入文件"), win)
			return
		}
		if current.output == "" {
			dialog.ShowError(errors.New("请先选择输出文件"), win)
			return
		}

		args := buildFFmpegArgs(current)
		logOutput.SetText("")
		statusLabel.SetText("正在调用 FFmpeg...")
		runButton.Disable()
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
				runButton.Enable()
				cancelButton.Disable()
			})

			if errors.Is(ctx.Err(), context.Canceled) {
				setStatus(statusLabel, "已停止")
				return
			}
			if err != nil {
				setStatus(statusLabel, "执行失败："+err.Error())
				return
			}
			setStatus(statusLabel, "处理完成："+filepath.Base(current.output))
		}()
	}

	title := canvas.NewText("HexImg", ui.TextColor(true))
	title.TextSize = 30
	title.TextStyle = fyne.TextStyle{Bold: true}
	subtitle := canvas.NewText(fmt.Sprintf("FFmpeg native desktop · Go + Fyne · %s · %s/%s", version, runtime.GOOS, runtime.GOARCH), ui.MutedTextColor(true))
	subtitle.TextSize = 13

	header := container.NewBorder(nil, nil, nil, themeButton, container.NewVBox(title, subtitle))

	sourceCard := fluentCard(true, "源文件", "选择输入和输出位置", container.NewVBox(
		widget.NewLabel("输入文件"),
		container.NewBorder(nil, nil, nil, openInputButton, inputEntry),
		widget.NewLabel("输出文件"),
		container.NewBorder(nil, nil, nil, openOutputButton, outputEntry),
	))

	settingsCard := fluentCard(true, "编码设置", "常用 FFmpeg 参数", container.NewVBox(
		container.NewGridWithColumns(3,
			container.NewVBox(widget.NewLabel("格式"), formatSelect),
			container.NewVBox(widget.NewLabel("质量"), qualitySelect),
			container.NewVBox(widget.NewLabel("预设"), presetSelect),
		),
		widget.NewLabel("缩放"),
		scaleEntry,
		widget.NewLabel("额外参数"),
		extraArgsEntry,
		overwriteCheck,
	))

	commandCard := fluentCard(true, "命令预览", "执行前可检查参数", commandPreview)
	logCard := fluentCard(true, "运行日志", "FFmpeg 输出", logOutput)

	actionBar := container.NewBorder(nil, nil, statusLabel, container.NewHBox(cancelButton, runButton), nil)
	mainPanel := container.NewBorder(header, actionBar, nil, nil, container.NewVScroll(container.NewVBox(
		sourceCard,
		settingsCard,
		commandCard,
		logCard,
	)))

	content := container.NewBorder(nil, nil, referencePanel(), nil, mainPanel)
	win.SetContent(content)
	win.ShowAndRun()
}

func referencePanel() fyne.CanvasObject {
	darkCard := paletteCard(true)
	lightCard := paletteCard(false)
	return container.NewVBox(
		darkCard,
		lightCard,
	)
}

func paletteCard(dark bool) fyne.CanvasObject {
	title := "fluent-light"
	if dark {
		title = "fluent-dark"
	}

	heading := canvas.NewText(title, ui.TextColor(dark))
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.TextSize = 20

	card := container.NewVBox(
		heading,
		colorRow(dark, "背景", ui.BackgroundColor(dark), ui.MutedTextColor(dark), false),
		colorRow(dark, "卡片层", ui.CardColor(dark), ui.MutedTextColor(dark), false),
		colorRow(dark, "强调色", ui.AccentColor(dark), ui.AccentTextColor(dark), false),
		colorRow(dark, "文字", ui.TextColor(dark), ui.InvertedTextColor(dark), false),
		buttonPreview(dark),
		menuPreview(dark),
	)

	return fixedWidth(280, cardBackground(dark, card))
}

func colorRow(dark bool, label string, fill color.Color, textColor color.Color, selected bool) fyne.CanvasObject {
	left := canvas.NewText(label, textColor)
	left.TextSize = 14
	left.TextStyle = fyne.TextStyle{Bold: selected}
	right := canvas.NewText(ui.ColorHex(fill), textColor)
	right.TextSize = 13

	row := container.NewBorder(nil, nil, left, right)
	return rowBackground(fill, row)
}

func buttonPreview(dark bool) fyne.CanvasObject {
	text := canvas.NewText("Button", ui.TextColor(dark))
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 14
	return rowBackground(ui.ButtonColor(dark), container.NewCenter(text))
}

func menuPreview(dark bool) fyne.CanvasObject {
	item := canvas.NewText("Menu item", ui.TextColor(dark))
	item.TextSize = 14
	selected := canvas.NewText("Selected", ui.TextColor(dark))
	selected.TextSize = 13

	selectedBox := rowBackground(ui.SelectionColor(dark), container.NewBorder(nil, nil, selected, nil))
	return rowBackground(ui.CardColor(dark), container.NewVBox(item, selectedBox))
}

func fluentCard(dark bool, titleText, subtitleText string, content fyne.CanvasObject) fyne.CanvasObject {
	title := canvas.NewText(titleText, ui.TextColor(dark))
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 18
	subtitle := canvas.NewText(subtitleText, ui.MutedTextColor(dark))
	subtitle.TextSize = 12

	body := container.NewVBox(title, subtitle, widget.NewSeparator(), content)
	return cardBackground(dark, body)
}

func cardBackground(dark bool, content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(ui.CardColor(dark))
	bg.CornerRadius = 14
	return container.NewStack(bg, container.NewPadded(content))
}

func rowBackground(fill color.Color, content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(fill)
	bg.CornerRadius = 6
	return container.NewStack(bg, container.NewPadded(content))
}

func fixedWidth(width float32, content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWrap(fyne.NewSize(width, content.MinSize().Height), content)
}

func icon(resource fyne.Resource) fyne.Resource {
	return theme.NewThemedResource(resource)
}

func buildFFmpegArgs(cfg encodeConfig) []string {
	args := []string{"-hide_banner"}
	if cfg.overwrite {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}
	if cfg.input != "" {
		args = append(args, "-i", cfg.input)
	}

	switch cfg.format {
	case "mp3":
		args = append(args, "-vn", "-c:a", "libmp3lame")
		args = append(args, audioQualityArgs(cfg.quality)...)
	case "wav":
		args = append(args, "-vn", "-c:a", "pcm_s16le")
	case "webm":
		args = append(args, "-c:v", "libvpx-vp9", "-c:a", "libopus")
		args = append(args, videoQualityArgs(cfg.quality)...)
	case "gif":
		filter := "fps=15"
		if cfg.scale != "" {
			filter += ",scale=" + cfg.scale
		}
		args = append(args, "-vf", filter)
	default:
		args = append(args, "-c:v", "libx264", "-c:a", "aac")
		if cfg.preset != "" && cfg.preset != "auto" {
			args = append(args, "-preset", cfg.preset)
		}
		args = append(args, videoQualityArgs(cfg.quality)...)
	}

	if cfg.scale != "" && cfg.format != "gif" {
		args = append(args, "-vf", "scale="+cfg.scale)
	}

	args = append(args, splitExtraArgs(cfg.extraArgs)...)
	if cfg.output != "" {
		args = append(args, cfg.output)
	}
	return args
}

func videoQualityArgs(quality string) []string {
	switch quality {
	case "高质量":
		return []string{"-crf", "18"}
	case "体积优先":
		return []string{"-crf", "30"}
	default:
		return []string{"-crf", "23"}
	}
}

func audioQualityArgs(quality string) []string {
	switch quality {
	case "高质量":
		return []string{"-b:a", "320k"}
	case "体积优先":
		return []string{"-b:a", "128k"}
	default:
		return []string{"-b:a", "192k"}
	}
}

func splitExtraArgs(input string) []string {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false

	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\n':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func shellPreview(binary string, args []string) string {
	quoted := make([]string, 0, len(args)+1)
	quoted = append(quoted, binary)
	for _, arg := range args {
		quoted = append(quoted, quoteArg(arg))
	}
	return strings.Join(quoted, " ")
}

func quoteArg(arg string) string {
	if arg == "" {
		return `""`
	}
	if strings.ContainsAny(arg, " \t\n\"'") {
		return `"` + strings.ReplaceAll(arg, `"`, `\"`) + `"`
	}
	return arg
}

func defaultOutputPath(inputPath, format string) string {
	ext := "." + format
	if format == "" {
		ext = ".mp4"
	}
	base := strings.TrimSuffix(inputPath, filepath.Ext(inputPath))
	return base + "_heximg" + ext
}

func runFFmpeg(ctx context.Context, args []string, appendLine func(string)) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

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
	readPipe := func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			appendLine(scanner.Text())
		}
	}
	wg.Add(2)
	go readPipe(stdout)
	go readPipe(stderr)

	waitErr := cmd.Wait()
	wg.Wait()
	return waitErr
}

func appendLog(logOutput *widget.Entry, line string) {
	fyne.Do(func() {
		const maxLogLength = 48000
		current := logOutput.Text
		if len(current) > maxLogLength {
			current = current[len(current)-maxLogLength:]
		}
		logOutput.SetText(current + line + "\n")
	})
}

func setStatus(label *widget.Label, text string) {
	fyne.Do(func() {
		label.SetText(text)
	})
}
