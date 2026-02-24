package main

import "testing"

func TestNativeManifestPathRejectsUnknownBrowser(t *testing.T) {
	_, err := nativeManifestPath("unknown")
	if err == nil {
		t.Fatal("expected error for unknown browser")
	}
}
