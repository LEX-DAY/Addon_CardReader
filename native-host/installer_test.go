package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNativeManifestPathRejectsUnknownBrowser(t *testing.T) {
	_, err := nativeManifestPath("unknown")
	if err == nil {
		t.Fatal("expected error for unknown browser")
	}
}

func TestNativeManifestPathSupportsEdge(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("edge manifest path is only defined for windows/linux")
	}

	p, err := nativeManifestPath("edge")
	if err != nil {
		t.Fatalf("expected edge to be supported, got error: %v", err)
	}
	if filepath.Ext(p) != ".json" {
		t.Fatalf("expected .json manifest path, got: %s", p)
	}
	if !strings.Contains(strings.ToLower(p), "edge") {
		t.Fatalf("expected edge-specific path, got: %s", p)
	}
}

func TestNativeManifestPathSupportsYandex(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		t.Skip("yandex manifest path is only defined for windows/linux")
	}

	p, err := nativeManifestPath("yandex")
	if err != nil {
		t.Fatalf("expected yandex to be supported, got error: %v", err)
	}
	if filepath.Ext(p) != ".json" {
		t.Fatalf("expected .json manifest path, got: %s", p)
	}
	if !strings.Contains(strings.ToLower(p), "yandex") {
		t.Fatalf("expected yandex-specific path, got: %s", p)
	}
}

func TestNormalizeBrowserAliases(t *testing.T) {
	cases := map[string]string{
		"chrome":         "chrome",
		" msedge ":       "edge",
		"Microsoft-Edge": "edge",
		"yabrowser":      "yandex",
		"YandexBrowser":  "yandex",
	}
	for in, want := range cases {
		got := normalizeBrowser(in)
		if got != want {
			t.Fatalf("normalizeBrowser(%q)=%q, want %q", in, got, want)
		}
	}
}
