package ffmpeg

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

var baseDir string

var dbgMode bool = false

//FFMpeg Contains FFMpeg's Executable Path and If it uses Preinstalled Executable
type FFMpeg struct {
	Executable            string
	UsePreinstalledFFMpeg bool
}

//Init Initializes FFMpeg's Fields and Downloads ffmpeg-prebuilt binaries
func (f *FFMpeg) Init() error {
	if f.UsePreinstalledFFMpeg {
		f.Executable = "ffmpeg"
		return nil
	}

	baseDir = "./ffmpeg-prebuilt"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		_, err := f.ExecWithDefaultHandle(
			"git", "clone", "https://github.com/sam1677/ffmpeg-prebuilt.git",
		)
		if err != nil {
			return err
		}
	}
	f.Executable = baseDir

	switch runtime.GOOS {
	case "windows":
		f.Executable += "/windows/ffmpeg.exe"

	case "linux":
		f.Executable += "/linux"

		switch runtime.GOARCH {
		case "386":
			f.Executable += "/386"
		case "amd64":
			f.Executable += "/amd64"
		case "arm":
			f.Executable += "/armhf"
		case "arm64":
			f.Executable += "/arm64"
		}
		f.Executable += "/ffmpeg"
	}

	return nil
}

//MergeVideoNAudio Merges a Video and a Audio into one Video
func (f *FFMpeg) MergeVideoNAudio(video *os.File, audio *os.File, path, outputFileName string) error {
	fmt.Println("Start merging video and audio")
	defer fmt.Println("End merging video and audio")

	err := u.CreatePath(path)
	if err != nil {
		return e.DbgErr(err)
	}

	_, err = f.ExecWithDefaultHandle(
		//		ffmpeg -i video.mp4 -i audio.mp3 -c:v copy -c:a copy -map 0:v -map 1:a -nostdin -y output.mp4
		fmt.Sprintf("%s -i %s -i %s -c:v copy -c:a copy -map 0:v -map 1:a -nostdin -y %s",
			f.Executable, video.Name(), audio.Name(), u.MergePathAndFilename(path, outputFileName),
		))

	if err != nil {
		return err
	}

	return nil
}

//Exec executes command
//It is different from *os.Command
func (f *FFMpeg) Exec(args ...string) (cmd *exec.Cmd, stdout <-chan []byte, stderr <-chan []byte, err error) {
	cmd = exec.Command("/bin/bash", "-c", strings.Join(args, " "))
	dbgPrintln(cmd)

	sout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, e.DbgErr(err)
	}

	serr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, e.DbgErr(err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, nil, e.DbgErr(err)
	}

	return cmd, NewChanFromReader(sout), NewChanFromReader(serr), nil
}

//ExecWithHandle Executes shell command with stdout and stderr handler functions
func (f *FFMpeg) ExecWithHandle(
	stdoutHandler func([]byte) error,
	stderrHandler func([]byte) error,
	deferFunc func(*os.ProcessState, string) error, args ...string) (*exec.Cmd, error) {

	cmd, stdout, stderr, err := f.Exec(args...)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	breaks := make(chan *os.ProcessState)
	lastError := ""

	go func() {
		state, _ := cmd.Process.Wait()
		breaks <- state
	}()

	for {
		select {
		case state := <-breaks:
			err = deferFunc(state, lastError)
			if err != nil {
				return nil, err
			}

			return cmd, nil
		case sout := <-stdout:
			err = stdoutHandler(sout)
		case serr := <-stderr:
			lastError = string(serr)
			err = stderrHandler(serr)
		}
		if err != nil {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			return cmd, e.DbgErr(err)
		}
	}
}

//ExecWithDefaultHandle Executes shell command with default stdout and stderr handler functions
func (f *FFMpeg) ExecWithDefaultHandle(args ...string) (*exec.Cmd, error) {
	cmd, err := f.ExecWithHandle(
		defaultStdoutHandler,
		defaultStderrHandler,
		defaultDeferFunc,
		args...,
	)
	return cmd, err
}

//NewChanFromReader creates goroutine which reads line from given io.Reader and sends them to chan
func NewChanFromReader(r io.ReadCloser) <-chan []byte {
	ch := make(chan []byte)
	go func() {
		buf := bufio.NewReader(r)
		defer r.Close()
		defer close(ch)
		for {
			byt, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			if len(byt) == 0 {
				continue
			}

			ch <- byt
		}

	}()
	return ch
}

func defaultStdoutHandler(sout []byte) error {
	out := strings.TrimSpace(string(sout))
	if out == "" {
		return nil
	}
	dbgPrintln(out)
	return nil
}

func defaultStderrHandler(serr []byte) error {
	err := strings.TrimSpace(string(serr))
	if err == "" {
		return nil
	}
	dbgPrintln(err)
	return nil
}

func defaultDeferFunc(state *os.ProcessState, lastError string) error {
	code := state.ExitCode()
	if state.Exited() && code != 0 {
		fmt.Println("code :", code)
		return e.DbgErr(errors.New(lastError))
	}
	return nil
}

func dbgPrintln(args ...interface{}) {
	if dbgMode {
		fmt.Println(args...)
	}
}
