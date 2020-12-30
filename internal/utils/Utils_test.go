package utils_test

import (
	"testing"

	"github.com/sam1677/ytdl"
)

func TestContentType(t *testing.T) {
	a := ytdl.Audio
	b := ytdl.Video | ytdl.VandA
	t.Log(a & ytdl.Video)
	t.Log(b & ytdl.Video)
}
