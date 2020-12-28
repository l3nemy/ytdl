package ytdl_test

import (
	"fmt"
	"testing"

	"github.com/sam1677/ytdl"
)

var vi *ytdl.VideoInfo
var err error

func TestGetVideoInfo(t *testing.T) {
	vi, err = ytdl.GetVideoInfo("https://www.youtube.com/watch?v=BL70VjXfFCk")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v\n", vi.CombinedFormatList().Sort())
}

func TestDownload(t *testing.T) {
	fmt.Println("getting videoInfo")
	if !t.Run("TestGetVideoInfo", TestGetVideoInfo) {
		panic(err)
	}
	fmt.Println("getting videoInfo done")

	audios := []*ytdl.Format{}
	for _, format := range vi.StreamingData.AdaptiveFormats {
		if format.Type == "audio" {
			audios = append(audios, format)
		}
	}

	err = vi.StreamingData.AdaptiveFormats[0].Download(&ytdl.DownloadOptions{AudioOverride: audios[0]})
	if err != nil {
		t.Error(err)
		return
	}
}
