package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestClampQuality(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "below range", in: -1, want: 0},
		{name: "in range", in: 85, want: 85},
		{name: "above range", in: 101, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampQuality(tt.in); got != tt.want {
				t.Fatalf("clampQuality(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestJpegQScale(t *testing.T) {
	tests := []struct {
		quality int
		want    int
	}{
		{quality: 0, want: 31},
		{quality: 85, want: 7},
		{quality: 100, want: 2},
	}

	for _, tt := range tests {
		if got := jpegQScale(tt.quality); got != tt.want {
			t.Fatalf("jpegQScale(%d) = %d, want %d", tt.quality, got, tt.want)
		}
	}
}

func TestPngCompressionLevel(t *testing.T) {
	tests := []struct {
		value int
		want  int
	}{
		{value: -1, want: 0},
		{value: 6, want: 6},
		{value: 10, want: 9},
	}

	for _, tt := range tests {
		if got := pngCompressionLevel(tt.value); got != tt.want {
			t.Fatalf("pngCompressionLevel(%d) = %d, want %d", tt.value, got, tt.want)
		}
	}
}

func TestBuildFFmpegArgsForPNGUsesCompressionLevel(t *testing.T) {
	got := buildFFmpegArgs(convertConfig{
		input:      "input.png",
		output:     "final.png",
		workOutput: "output.png",
		format:     "png",
		quality:    6,
	})
	want := []string{
		"-hide_banner",
		"-y",
		"-i",
		"input.png",
		"-frames:v",
		"1",
		"-compression_level",
		"6",
		"output.png",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildFFmpegArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildFFmpegArgsForWebPClampsQuality(t *testing.T) {
	got := buildFFmpegArgs(convertConfig{
		input:   "input.png",
		output:  "output.webp",
		format:  "webp",
		quality: 120,
	})
	want := []string{
		"-hide_banner",
		"-y",
		"-i",
		"input.png",
		"-frames:v",
		"1",
		"-compression_level",
		"6",
		"-quality",
		"100",
		"output.webp",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildFFmpegArgs() = %#v, want %#v", got, want)
	}
}

func TestOutputForSuffixMode(t *testing.T) {
	got, replace := outputFor(filepath.Join("images", "photo.png"), "jpg", outputSettings{
		mode:   outputModeSuffix,
		suffix: defaultSuffix,
	})
	want := filepath.Join("images", "photo_HexImg.jpg")

	if got != want || replace {
		t.Fatalf("outputFor() = %q, %v, want %q, false", got, replace, want)
	}
}

func TestOutputForFolderModeSanitizesFolderName(t *testing.T) {
	got, replace := outputFor(filepath.Join("images", "photo.png"), "webp", outputSettings{
		mode:       outputModeFolder,
		folderName: `Hex:Img`,
	})
	want := filepath.Join("images", "Hex_Img", "photo.webp")

	if got != want || replace {
		t.Fatalf("outputFor() = %q, %v, want %q, false", got, replace, want)
	}
}

func TestOutputForReplaceMode(t *testing.T) {
	got, replace := outputFor(filepath.Join("images", "photo.png"), "png", outputSettings{
		mode: outputModeReplace,
	})
	want := filepath.Join("images", "photo.png")

	if got != want || !replace {
		t.Fatalf("outputFor() = %q, %v, want %q, true", got, replace, want)
	}
}

func TestPrepareOutputCreatesFolderForNonReplaceMode(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "nested", "photo.png")

	work, cleanup, err := prepareOutput(convertConfig{
		output: output,
		format: "png",
	})
	defer cleanup()
	if err != nil {
		t.Fatalf("prepareOutput() error = %v", err)
	}
	if work != output {
		t.Fatalf("work output = %q, want %q", work, output)
	}
	if info, err := os.Stat(filepath.Dir(output)); err != nil || !info.IsDir() {
		t.Fatalf("output dir stat = %v, info = %#v", err, info)
	}
}

func TestPrepareOutputCreatesReplaceTempInOutputDir(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "photo.jpg")
	output := filepath.Join(dir, "photo.png")

	if err := os.WriteFile(source, []byte("original jpg"), 0644); err != nil {
		t.Fatal(err)
	}

	work, cleanup, err := prepareOutput(convertConfig{
		input:         source,
		output:        output,
		format:        "png",
		replaceSource: true,
	})
	defer cleanup()
	if err != nil {
		t.Fatalf("prepareOutput() error = %v", err)
	}
	if filepath.Dir(work) != dir {
		t.Fatalf("work output dir = %q, want %q", filepath.Dir(work), dir)
	}
	if _, err := os.Stat(work); err != nil {
		t.Fatalf("work output stat error = %v", err)
	}
}

func TestFinalizeOutputReplacesSamePathThroughBackup(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "photo.png")
	work := filepath.Join(dir, ".heximg-work.png")

	if err := os.WriteFile(source, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(work, []byte("converted"), 0644); err != nil {
		t.Fatal(err)
	}

	err := finalizeOutput(convertConfig{
		input:         source,
		output:        source,
		format:        "png",
		replaceSource: true,
		workOutput:    work,
	})
	if err != nil {
		t.Fatalf("finalizeOutput() error = %v", err)
	}

	got, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "converted" {
		t.Fatalf("source content = %q, want converted", got)
	}
	if _, err := os.Stat(work); !os.IsNotExist(err) {
		t.Fatalf("work output still exists, stat err = %v", err)
	}
}

func TestFinalizeOutputMovesDifferentPathAndRemovesInput(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "photo.jpg")
	output := filepath.Join(dir, "photo.png")
	work := filepath.Join(dir, ".heximg-work.png")

	if err := os.WriteFile(source, []byte("original jpg"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(work, []byte("converted png"), 0644); err != nil {
		t.Fatal(err)
	}

	err := finalizeOutput(convertConfig{
		input:         source,
		output:        output,
		format:        "png",
		replaceSource: true,
		workOutput:    work,
	})
	if err != nil {
		t.Fatalf("finalizeOutput() error = %v", err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "converted png" {
		t.Fatalf("output content = %q, want converted png", got)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists, stat err = %v", err)
	}
}

func TestPrepareOutputRefusesExistingReplaceTarget(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "photo.jpg")
	output := filepath.Join(dir, "photo.png")

	if err := os.WriteFile(input, []byte("original jpg"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("existing png"), 0644); err != nil {
		t.Fatal(err)
	}

	_, cleanup, err := prepareOutput(convertConfig{
		input:         input,
		output:        output,
		format:        "png",
		replaceSource: true,
	})
	cleanup()
	if err == nil {
		t.Fatal("prepareOutput() error = nil, want existing target error")
	}

	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "existing png" {
		t.Fatalf("existing output content = %q, want unchanged", got)
	}
}

func TestSanitizePathPart(t *testing.T) {
	if got := sanitizePathPart("  ", defaultFolderName); got != defaultFolderName {
		t.Fatalf("sanitizePathPart(blank) = %q, want %q", got, defaultFolderName)
	}
	if got := sanitizePathPart(`A:B/C\D*E?F"G<H>I|J`, defaultFolderName); got != "A_B_C_D_E_F_G_H_I_J" {
		t.Fatalf("sanitizePathPart(invalid) = %q", got)
	}
}

func TestSamePathCleansEquivalentPaths(t *testing.T) {
	dir := t.TempDir()
	left := filepath.Join(dir, "nested", "..", "photo.png")
	right := filepath.Join(dir, "photo.png")

	if !samePath(left, right) {
		t.Fatalf("samePath(%q, %q) = false, want true", left, right)
	}
}
