package main

import (
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
		input:   "input.png",
		output:  "output.png",
		format:  "png",
		quality: 6,
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
