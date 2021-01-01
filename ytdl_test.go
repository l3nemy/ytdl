package ytdl_test

import (
	"fmt"
	"testing"

	"github.com/sam1677/ytdl"
)

var vi *ytdl.VideoInfo
var err error

func TestGetVideoInfo(t *testing.T) {
	vi, err = ytdl.GetVideoInfo("https://www.youtube.com/watch?v=9bZkp7q19f0")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v\n", vi.CombinedFormatList().Sort())
}

func TestDownload(t *testing.T) {
	fmt.Println("getting videoInfo")
	if vi == nil {
		if !t.Run("TestGetVideoInfo", TestGetVideoInfo) {
			panic(err)
		}
	}

	fmt.Println("getting videoInfo done")

	audios := vi.StreamingData.AdaptiveFormats.Audios()

	_, err = vi.StreamingData.AdaptiveFormats.Videos().Best().Download(&ytdl.DownloadOptions{AudioOverride: audios.Best()})
	if err != nil {
		t.Error(err)
		return
	}
}
