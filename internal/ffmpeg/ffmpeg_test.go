package ffmpeg_test

import (
	"os"
	"testing"

	"github.com/sam1677/ytdl/internal/ffmpeg"
)

func TestInit(t *testing.T) {
	err := os.RemoveAll("./ffmpeg-prebuilt")
	if err != nil {
		t.Error(err)
		return
	}
	f := new(ffmpeg.FFMpeg)
	err = f.Init()
	if err != nil {
		t.Error(err)
		return
	}
}
