package ytdl

import (
	"os"

	"github.com/sam1677/ytdl/internal/ffmpeg"
	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

const getVideoInfoURL = "https://www.youtube.com/get_video_info?video_id=%s"

const (
	tmpJSONDir          = "./jsonCache"
	tmpAudioDir         = "./audioCache"
	tmpVideoDir         = "./videoCache"
	tmpScriptDir        = "./scriptCache"
	downloadDefaultPath = "./Downloads"
)

//DownloadOptions contains Download Path, Filename
//and AudioFormat
type DownloadOptions struct {
	Path          string
	Filename      string
	AudioOverride *Format
}

//Download Downloads format and overrides audio if AudioOverride is not nil
func (f *Format) Download(options *DownloadOptions) error {
	if options == nil {
		options = new(DownloadOptions)
	}
	if options.Path == "" {
		options.Path = downloadDefaultPath
	}
	if options.Filename == "" {
		options.Filename = f.Filename
	}
	path := options.Path
	if options.AudioOverride != nil {
		path = tmpVideoDir
	}

	file, err := f.downloadWithPath(path, options.Filename)
	if err != nil {
		return e.DbgErr(err)
	}
	defer file.Close()

	if options.AudioOverride != nil {
		err = f.audioOverride(options.AudioOverride, file, options.Path)
		if err != nil {
			return e.DbgErr(err)
		}
	}
	return nil
}

func (f *Format) downloadWithPath(path string, filename string) (*os.File, error) {
	if f.URL == "" {
		return nil, e.DbgErr(e.ErrURLIsEmpty)
	}

	file, _, err := u.DownloadFile(f.URL, path, filename, false)
	if err != nil {
		return nil, e.DbgErr(err)
	}
	return file, nil
}

func (f *Format) audioOverride(audio *Format, videoFile *os.File, finalDir string) error {
	audioFile, err := audio.downloadWithPath(tmpAudioDir, audio.Filename)
	if err != nil {
		return e.DbgErr(err)
	}

	ff := new(ffmpeg.FFMpeg)

	err = ff.Init()
	if err != nil {
		return e.DbgErr(err)
	}

	vinfo, err := videoFile.Stat()
	if err != nil {
		return err
	}

	err = ff.MergeVideoNAudio(videoFile, audioFile, finalDir, vinfo.Name())
	if err != nil {
		return e.DbgErr(err)
	}

	audioFile.Close()
	return nil
}
