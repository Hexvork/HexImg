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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fas "github.com/Hexvork/HexImg/internal/fas"
	"github.com/Hexvork/HexImg/internal/ui"
)

var version = "dev"

type convertConfig struct {
	input         string
	output        string
	format        string
	quality       int
	replaceSource bool
	workOutput    string
}

type outputSettings struct {
	mode       string
	suffix     string
	folderName string
}

const (
	outputModeSuffix  = "添加后缀"
	outputModeFolder  = "当前目录文件夹"
	outputModeReplace = "替换源文件"
	defaultSuffix     = "_HexImg"
	defaultFolderName = "HexImg"
)

func main() {
	hex := app.NewWithID("com.hexvork.heximg")
	darkMode := true
	hex.Settings().SetTheme(ui.NewFluentTheme(darkMode))

	win := hex.NewWindow("HexImg")
	win.SetIcon(fas.Image)
	win.Resize(fyne.NewSize(760, 520))
	win.SetMaster()

	// Image list: each file is a block with a delete button
	var imagePaths []string

	imageChips := container.NewHBox()
	imageChipsScroll := container.NewScroll(imageChips)
	imageChipsScroll.SetMinSize(fyne.NewSize(0, 42))
	imageChipsScroll.Hide()

	// Forward-declared; assigned after all helpers are defined below
	var clearAllBtn *widget.Button
	var refreshImageChips func()
	var refreshDropHint func()

	allInputs := func() []string { return imagePaths }

	outputPath := widget.NewEntry()
	outputPath.SetPlaceHolder("输出路径会自动生成")
	outputPath.Disable()

	statusLabel := widget.NewLabel("就绪")
	statusLabel.Wrapping = fyne.TextWrapWord
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		statusLabel.SetText("未检测到 FFmpeg，请先安装并加入 PATH")
	}

	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.SetMinRowsVisible(5)
	logOutput.Disable()

	formatSelect := widget.NewRadioGroup([]string{"jpg", "png", "webp", "bmp", "tiff"}, nil)
	formatSelect.Horizontal = true
	formatSelect.SetSelected("jpg")

	outputModeSelect := widget.NewSelect([]string{outputModeSuffix, outputModeFolder, outputModeReplace}, nil)
	outputModeSelect.SetSelected(outputModeSuffix)

	suffixEntry := widget.NewEntry()
	suffixEntry.SetText(defaultSuffix)

	folderEntry := widget.NewEntry()
	folderEntry.SetText(defaultFolderName)
	folderEntry.Disable()

	qualityLabel := widget.NewLabel("质量")
	qualityValue := widget.NewLabel("85")
	qualitySlider := widget.NewSlider(0, 100)
	qualitySlider.Step = 1
	qualitySlider.Value = 85
	qualitySlider.OnChanged = func(value float64) {
		qualityValue.SetText(strconv.Itoa(int(value)))
	}

	cfg := func() convertConfig {
		var first string
		if len(imagePaths) > 0 {
			first = imagePaths[0]
		}
		format := formatSelect.Selected
		settings := outputSettings{
			mode:       outputModeSelect.Selected,
			suffix:     strings.TrimSpace(suffixEntry.Text),
			folderName: strings.TrimSpace(folderEntry.Text),
		}
		output, replaceSource := outputFor(first, format, settings)
		return convertConfig{
			input:         first,
			output:        output,
			format:        format,
			quality:       int(qualitySlider.Value),
			replaceSource: replaceSource,
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

	refreshOutputMode := func(mode string) {
		suffixEntry.Disable()
		folderEntry.Disable()
		switch mode {
		case outputModeFolder:
			folderEntry.Enable()
		case outputModeReplace:
		default:
			suffixEntry.Enable()
		}
		refreshOutput()
	}
	outputModeSelect.OnChanged = refreshOutputMode
	suffixEntry.OnChanged = func(string) {
		refreshOutput()
	}
	folderEntry.OnChanged = func(string) {
		refreshOutput()
	}

	// Hint shown when no image has been selected yet
	dropHint := widget.NewLabelWithStyle("可拖拽图片至窗口", fyne.TextAlignCenter, fyne.TextStyle{})
	dropHint.Importance = widget.LowImportance
	refreshDropHint = func() {
		if len(imagePaths) == 0 {
			dropHint.Show()
		} else {
			dropHint.Hide()
		}
	}

	openInputButton := widget.NewButtonWithIcon("选择图片", icon(fas.FolderOpen), func() {
		chooseImageFile(win, func(paths []string, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if len(paths) == 0 {
				return
			}
			imagePaths = append(imagePaths, paths...)
			refreshImageChips()
			refreshOutput()
			refreshDropHint()
			if len(imagePaths) == 1 {
				statusLabel.SetText("已选择：" + filepath.Base(imagePaths[0]))
			} else {
				statusLabel.SetText(fmt.Sprintf("已选择 %d 张图片", len(imagePaths)))
			}
		})
	})

	var cancelMu sync.Mutex
	var cancelRun context.CancelFunc

	convertButton := widget.NewButtonWithIcon("转换", icon(fas.Play), nil)
	convertButton.Importance = widget.HighImportance

	cancelButton := widget.NewButtonWithIcon("停止", icon(fas.Stop), func() {
		cancelMu.Lock()
		defer cancelMu.Unlock()
		if cancelRun != nil {
			cancelRun()
			statusLabel.SetText("正在停止转换...")
		}
	})
	cancelButton.Disable()

	convertButton.OnTapped = func() {
		inputs := allInputs()
		if len(inputs) == 0 {
			dialog.ShowError(errors.New("请先选择图片"), win)
			return
		}
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			dialog.ShowError(errors.New("未找到 ffmpeg，请先安装 FFmpeg 并加入 PATH"), win)
			statusLabel.SetText("未检测到 FFmpeg，请先安装并加入 PATH")
			return
		}

		format := formatSelect.Selected
		if format == "" {
			dialog.ShowError(errors.New("请先选择转换格式"), win)
			return
		}

		quality := int(qualitySlider.Value)
		settings := outputSettings{
			mode:       outputModeSelect.Selected,
			suffix:     strings.TrimSpace(suffixEntry.Text),
			folderName: strings.TrimSpace(folderEntry.Text),
		}

		logOutput.SetText("")
		statusLabel.SetText("正在转换...")
		convertButton.Disable()
		cancelButton.Enable()

		ctx, cancel := context.WithCancel(context.Background())
		cancelMu.Lock()
		cancelRun = cancel
		cancelMu.Unlock()

		go func() {
			total := len(inputs)
			for i, input := range inputs {
				select {
				case <-ctx.Done():
					fyne.Do(func() {
						statusLabel.SetText("已停止")
					})
					cancelMu.Lock()
					cancelRun = nil
					cancelMu.Unlock()
					fyne.Do(func() {
						convertButton.Enable()
						cancelButton.Disable()
					})
					return
				default:
				}

				output, replaceSource := outputFor(input, format, settings)
				cfg := convertConfig{
					input:         input,
					output:        output,
					format:        format,
					quality:       quality,
					replaceSource: replaceSource,
				}

				workOutput, cleanup, err := prepareOutput(cfg)
				if err != nil {
					fyne.Do(func() {
						appendLog(logOutput, fmt.Sprintf("[%d/%d] %s 准备失败：%s", i+1, total, filepath.Base(input), err))
					})
					cleanup()
					continue
				}
				cfg.workOutput = workOutput

				args := buildFFmpegArgs(cfg)
				err = runFFmpeg(ctx, args, func(line string) {
					appendLog(logOutput, line)
				})
				cleanup()

				if err != nil {
					fyne.Do(func() {
						appendLog(logOutput, fmt.Sprintf("[%d/%d] %s 转换失败：%s", i+1, total, filepath.Base(input), err))
					})
					continue
				}

				if err := finalizeOutput(cfg); err != nil {
					fyne.Do(func() {
						appendLog(logOutput, fmt.Sprintf("[%d/%d] %s 保存失败：%s", i+1, total, filepath.Base(input), err))
					})
					continue
				}

				fyne.Do(func() {
					appendLog(logOutput, fmt.Sprintf("[%d/%d] ✓ %s", i+1, total, filepath.Base(output)))
				})
			}

			cancelMu.Lock()
			cancelRun = nil
			cancelMu.Unlock()
			fyne.Do(func() {
				convertButton.Enable()
				cancelButton.Disable()
				statusLabel.SetText("转换完成")
			})
		}()
	}

	clearAllBtn = widget.NewButtonWithIcon("清空", icon(fas.Stop), func() {
		imagePaths = nil
		refreshImageChips()
		refreshOutput()
		refreshDropHint()
		statusLabel.SetText("就绪")
	})
	clearAllBtn.Importance = widget.LowImportance
	clearAllBtn.Hide()

	refreshImageChips = func() {
		imageChips.RemoveAll()
		if len(imagePaths) == 0 {
			imageChipsScroll.Hide()
			clearAllBtn.Hide()
			return
		}
		imageChipsScroll.Show()
		clearAllBtn.Show()
		for idx, path := range imagePaths {
			i, p := idx, path
			name := filepath.Base(p)
			removeBtn := widget.NewButton("×", func() {
				imagePaths = append(imagePaths[:i], imagePaths[i+1:]...)
				refreshImageChips()
				refreshOutput()
				refreshDropHint()
				if len(imagePaths) == 0 {
					statusLabel.SetText("就绪")
				}
			})
			removeBtn.Importance = widget.LowImportance
			chip := container.NewHBox(
				widget.NewLabel(name),
				removeBtn,
			)
			imageChips.Add(chip)
		}
		imageChips.Refresh()
	}

	title := widget.NewLabelWithStyle("HexImg", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitle := widget.NewLabel(fmt.Sprintf("图片格式转换 · FFmpeg · %s · %s/%s", version, runtime.GOOS, runtime.GOARCH))
	subtitle.Importance = widget.LowImportance

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
	})
	themeButton.Importance = widget.LowImportance

	// Square rounded rectangle for the theme toggle
	themeButtonContainer := container.NewGridWrap(fyne.NewSize(36, 36), themeButton)

	header := container.NewBorder(nil, nil, nil, nil,
		container.NewHBox(
			container.NewVBox(title, subtitle),
			layout.NewSpacer(),
			themeButtonContainer,
		),
	)

	// Accept dragged-in image files (append to existing, no duplicates)
	win.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		existingSet := make(map[string]bool, len(imagePaths))
		for _, p := range imagePaths {
			existingSet[p] = true
		}
		var added int
		for _, uri := range uris {
			path := uri.Path()
			if isImageFile(path) && !existingSet[path] {
				imagePaths = append(imagePaths, path)
				existingSet[path] = true
				added++
			}
		}
		if added > 0 {
			refreshImageChips()
			refreshOutput()
			refreshDropHint()
			statusLabel.SetText(fmt.Sprintf("已选择 %d 张图片", len(imagePaths)))
		}
	})
	sourceCard := widget.NewCard("图片", "", container.NewVBox(
		widget.NewLabel("输入图片"),
		imageChipsScroll,
		container.NewBorder(nil, nil, openInputButton, clearAllBtn),
		dropHint,
		widget.NewLabel("输出文件"),
		outputPath,
	))

	settingsCard := widget.NewCard("转换设置", "", container.NewVBox(
		widget.NewLabel("目标格式"),
		formatSelect,
		container.NewBorder(nil, nil, qualityLabel, qualityValue, qualitySlider),
		widget.NewLabel("输出方式"),
		outputModeSelect,
		container.NewBorder(nil, nil, widget.NewLabel("后缀"), nil, suffixEntry),
		container.NewBorder(nil, nil, widget.NewLabel("文件夹"), nil, folderEntry),
	))

	logCard := widget.NewCard("状态", "", container.NewVBox(statusLabel, logOutput))
	actionBar := container.NewPadded(container.NewHBox(cancelButton, convertButton))
	mainContent := container.NewVScroll(container.NewPadded(container.NewVBox(sourceCard, settingsCard, logCard)))
	mainContent.SetMinSize(fyne.NewSize(720, 360))

	content := container.NewBorder(
		container.NewPadded(header),
		actionBar,
		nil,
		nil,
		mainContent,
	)
	win.SetContent(content)
	refreshOutput()
	refreshQualityControl(formatSelect.Selected)
	refreshOutputMode(outputModeSelect.Selected)
	refreshDropHint()
	win.ShowAndRun()
}

func icon(resource fyne.Resource) fyne.Resource {
	return theme.NewThemedResource(resource)
}

func outputFor(inputPath, format string, settings outputSettings) (string, bool) {
	if inputPath == "" {
		return "", false
	}
	if format == "" {
		format = "jpg"
	}
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	dir := filepath.Dir(inputPath)

	switch settings.mode {
	case outputModeFolder:
		folderName := sanitizePathPart(settings.folderName, defaultFolderName)
		return filepath.Join(dir, folderName, base+"."+format), false
	case outputModeReplace:
		return filepath.Join(dir, base+"."+format), true
	default:
		suffix := settings.suffix
		if suffix == "" {
			suffix = defaultSuffix
		}
		return filepath.Join(dir, base+suffix+"."+format), false
	}
}

func buildFFmpegArgs(cfg convertConfig) []string {
	args := []string{"-hide_banner", "-y", "-i", cfg.input}

	switch cfg.format {
	case "jpg":
		args = append(args, "-frames:v", "1", "-q:v", strconv.Itoa(jpegQScale(cfg.quality)))
	case "webp":
		args = append(args, "-frames:v", "1", "-compression_level", "6", "-quality", strconv.Itoa(clampQuality(cfg.quality)))
	case "png":
		args = append(args, "-frames:v", "1", "-compression_level", strconv.Itoa(pngCompressionLevel(cfg.quality)))
	case "bmp", "tiff":
		args = append(args, "-frames:v", "1")
	default:
		args = append(args, "-frames:v", "1")
	}

	output := cfg.output
	if cfg.workOutput != "" {
		output = cfg.workOutput
	}
	return append(args, output)
}

func prepareOutput(cfg convertConfig) (string, func(), error) {
	if cfg.output == "" {
		return "", func() {}, errors.New("输出路径不可用")
	}
	outputDir := filepath.Dir(cfg.output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", func() {}, fmt.Errorf("创建输出目录失败：%w", err)
	}
	if !cfg.replaceSource {
		return cfg.output, func() {}, nil
	}
	if !samePath(cfg.input, cfg.output) {
		if _, err := os.Stat(cfg.output); err == nil {
			return "", func() {}, fmt.Errorf("输出文件已存在，未覆盖：%s", cfg.output)
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", func() {}, fmt.Errorf("检查输出文件失败：%w", err)
		}
	}

	temp, err := os.CreateTemp(outputDir, ".heximg-*."+cfg.format)
	if err != nil {
		return "", func() {}, fmt.Errorf("创建临时输出文件失败：%w", err)
	}
	workOutput := temp.Name()
	if err := temp.Close(); err != nil {
		_ = os.Remove(workOutput)
		return "", func() {}, fmt.Errorf("关闭临时输出文件失败：%w", err)
	}
	cleanup := func() {
		_ = os.Remove(workOutput)
	}
	return workOutput, cleanup, nil
}

type finalizeWarning struct {
	message string
	err     error
}

func (w *finalizeWarning) Error() string {
	if w.err == nil {
		return w.message
	}
	return w.message + "：" + w.err.Error()
}

func (w *finalizeWarning) Unwrap() error {
	return w.err
}

func finalizeOutput(cfg convertConfig) error {
	if !cfg.replaceSource {
		return nil
	}
	if cfg.workOutput == "" {
		return errors.New("临时输出路径不可用")
	}
	if samePath(cfg.input, cfg.output) {
		backup, err := reserveBackupPath(cfg.output)
		if err != nil {
			return fmt.Errorf("创建源文件备份路径失败：%w", err)
		}
		if err := os.Rename(cfg.output, backup); err != nil {
			return fmt.Errorf("备份源文件失败：%w", err)
		}
		if err := moveFileNoOverwrite(cfg.workOutput, cfg.output); err != nil {
			if restoreErr := os.Rename(backup, cfg.output); restoreErr != nil {
				return fmt.Errorf("替换源文件失败：%w；恢复源文件也失败，备份保留在 %s：%v", err, backup, restoreErr)
			}
			return fmt.Errorf("替换源文件失败，已恢复原文件：%w", err)
		}
		if err := os.Remove(backup); err != nil {
			return &finalizeWarning{message: "清理备份失败；替换文件已保存", err: err}
		}
		return nil
	}
	if err := moveFileNoOverwrite(cfg.workOutput, cfg.output); err != nil {
		return fmt.Errorf("保存替换文件失败：%w", err)
	}
	if err := os.Remove(cfg.input); err != nil {
		return &finalizeWarning{message: "删除源文件失败；输出文件已保存", err: err}
	}
	return nil
}

func reserveBackupPath(target string) (string, error) {
	dir := filepath.Dir(target)
	ext := filepath.Ext(target)
	base := sanitizePathPart(strings.TrimSuffix(filepath.Base(target), ext), "source")
	temp, err := os.CreateTemp(dir, ".heximg-"+base+"-*.bak"+ext)
	if err != nil {
		return "", err
	}
	name := temp.Name()
	closeErr := temp.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		return "", closeErr
	}
	if removeErr != nil {
		return "", removeErr
	}
	return name, nil
}

func moveFileNoOverwrite(src, dst string) error {
	if err := os.Link(src, dst); err == nil {
		_ = os.Remove(src)
		return nil
	} else if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("目标文件已存在，未覆盖：%s", dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("目标文件已存在，未覆盖：%s", dst)
		}
		return err
	}

	cleanupDst := true
	defer func() {
		if cleanupDst {
			_ = os.Remove(dst)
		}
	}()
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return err
	}
	if err := dstFile.Close(); err != nil {
		return err
	}
	cleanupDst = false
	_ = os.Remove(src)
	return nil
}

func sanitizePathPart(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	replacer := strings.NewReplacer("\\", "_", "/", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(value)
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr == nil && rightErr == nil {
		left = leftAbs
		right = rightAbs
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
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
		logOutput.Append(line + "\n")
		if len(logOutput.Text) <= maxLogLength {
			return
		}
		runes := []rune(logOutput.Text)
		if len(runes) > maxLogLength {
			logOutput.SetText(string(runes[len(runes)-maxLogLength:]))
		}
	})
}

func setStatus(label *widget.Label, text string) {
	fyne.Do(func() {
		label.SetText(text)
	})
}

func isImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tif", ".tiff", ".gif":
		return true
	}
	return false
}
