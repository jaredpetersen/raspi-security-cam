package mpegdash

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

const fakeVideoStreamContent = "fakevideostream"

func TestMain(m *testing.M) {
	switch os.Getenv("GO_TEST_MODE") {
	case "":
		os.Exit(m.Run())
	case "ffmpeg":
		os.Stdout.WriteString(fakeVideoStreamContent)
		os.Exit(0)
	}
}

func TestStart(t *testing.T) {
	testCases := []struct {
		muxer        Muxer
		expectedArgs []string
	}{
		{
			Muxer{},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"livestream.mpd",
			},
		},
		{
			Muxer{Directory: "camera"},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				path.Join("camera", "livestream.mpd"),
			},
		},
		{
			Muxer{Fps: 60},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"-r", "60",
				"livestream.mpd",
			},
		},
		{
			Muxer{SegmentTime: 2},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"-seg_duration", "2",
				"livestream.mpd",
			},
		},
		{
			Muxer{PlaylistSize: 50},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"-window_size", "50",
				"livestream.mpd",
			},
		},
		{
			Muxer{StorageSize: 100},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"-extra_window_size", "100",
				"livestream.mpd",
			},
		},
		{
			Muxer{Directory: "mpegdash", Fps: 30, SegmentTime: 5, PlaylistSize: 25, StorageSize: 50},
			[]string{
				"ffmpeg",
				"-codec", "copy",
				"-f", "dash",
				"-re",
				"-an",
				"-init_seg_name", "init.m4s",
				"-media_seg_name", "$Time$-$Number$.m4s",
				"-r", "30",
				"-seg_duration", "5",
				"-window_size", "25",
				"-extra_window_size", "50",
				path.Join("mpegdash", "livestream.mpd"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.muxer), func(t *testing.T) {
			execCommand = mockExecCommand
			defer func() { execCommand = exec.Command }()

			videoStream := ioutil.NopCloser(strings.NewReader("totallyfakevideostream"))

			mpegdashMuxer := tc.muxer
			err := mpegdashMuxer.Start(videoStream)

			if err != nil {
				t.Error("Start produced an err", err)
			}

			ffmpegArgs := mpegdashMuxer.cmd.Args[1:]

			if !equal(ffmpegArgs, tc.expectedArgs) {
				t.Error("Command args do not match, got", ffmpegArgs)
			}
		})
	}
}

func TestStartReturnsError(t *testing.T) {
	execCommand = mockFailedExecCommand
	defer func() { execCommand = exec.Command }()

	videoStream := ioutil.NopCloser(strings.NewReader("totallyfakevideostream"))

	mpegdashMuxer := Muxer{}
	err := mpegdashMuxer.Start(videoStream)

	if err == nil {
		t.Error("Start failed to return an error")
	}
}

func TestWait(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	videoStream := ioutil.NopCloser(strings.NewReader("totallyfakevideostream"))

	mpegdashMuxer := Muxer{}
	mpegdashMuxer.Start(videoStream)
	err := mpegdashMuxer.Wait()

	if err != nil {
		t.Error("Wait returned an error", err)
	}
}

func TestWaitWithoutStartReturnsError(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	mpegdashMuxer := Muxer{}
	err := mpegdashMuxer.Wait()

	if err == nil || err.Error() != "ffmpeg mpegdash: not started" {
		t.Error("Wait failed to return correct error when run without Start", err)
	}
}

func TestWaitAgainReturnsError(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()

	videoStream := ioutil.NopCloser(strings.NewReader("totallyfakevideostream"))

	mpegdashMuxer := Muxer{}
	mpegdashMuxer.Start(videoStream)
	mpegdashMuxer.Wait()
	err := mpegdashMuxer.Wait()

	if err == nil {
		t.Error("Wait failed to return an error")
	}
}

func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := append([]string{command}, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(os.Environ(), "GO_TEST_MODE=ffmpeg")
	return cmd
}

func mockFailedExecCommand(command string, args ...string) *exec.Cmd {
	cmd := exec.Command("totallyfakecommandthatdoesnotexist")
	return cmd
}

func equal(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
