package magnet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_getSubtitleFiles(t *testing.T) {
	m := &Manager{}
	path := "testdata/[VCB-Studio] test"
	subtitleFiles, err := m.getSubtitleFiles(path)
	if err != nil {
		t.Fatalf("getSubtitleFiles failed: %v", err)
	}
	want := []string{
		"testdata/[VCB-Studio] test/a.ass",
	}
	assert.Equal(t, want, subtitleFiles)
}
