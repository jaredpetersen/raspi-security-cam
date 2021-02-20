package video

import (
	"log"
	"net/http"
	"sync"

	"github.com/jaredpetersen/raspilive/internal/ffmpeg/hls"
	"github.com/jaredpetersen/raspilive/internal/ffmpeg/mpegdash"
	"github.com/jaredpetersen/raspilive/internal/raspivid"
	"github.com/jaredpetersen/raspilive/internal/utils/pointer"
)

// ServeHls starts a static file server and stream video from the Raspberry Pi camera module using the HLS format.
//
// This is a blocking operation.
func ServeHls() {
	streamAndServe(streamHls)
}

// ServeMpegDash starts a static file server and stream video from the Raspberry Pi camera module using the HLS format.
//
// This is a blocking operation.
func ServeMpegDash() {
	streamAndServe(streamMpegDash)
}

// streamAndServe starts a static file server and streams video using the provided stream function.
//
// This is a blocking operation.
func streamAndServe(streamFn func()) {
	var wg sync.WaitGroup
	wg.Add(2)

	// Serve files generated by the video stream
	go func() {
		serveFiles()
		wg.Done()
	}()

	// Stream video
	go func() {
		streamFn()
		wg.Done()
	}()

	wg.Wait()
}

// StreamHls streams video from the Raspberry Pi camera module and muxes it to HLS.
//
// This is a blocking operation that will not complete.
func streamHls() {
	log.Println("Processing HLS")

	// Pipe video stream from raspivid into ffmpeg
	raspivid := raspivid.Stream(raspivid.Options{Width: 0})
	ffmpeg := hls.Hls(raspivid.Video, "./camera", hls.Options{Time: pointer.ToInt(0)})

	// Start ffmpeg first so that it's ready to accept the stream
	ffmpeg.Start()
	raspivid.Start()
	raspivid.Wait()
	ffmpeg.Wait()
}

// StreamMpegDash streams video from the Raspberry Pi camera module and muxes it to MPEG-DASH.
//
// This is a blocking operation that will not complete.
func streamMpegDash() {
	log.Println("Processing MPEG-DASH")

	// Pipe video stream from raspivid into ffmpeg
	raspivid := raspivid.Stream(raspivid.Options{Width: 0})
	ffmpeg := mpegdash.MpegDash(raspivid.Video, "./camera", mpegdash.Options{Time: pointer.ToInt(0)})

	// Start ffmpeg first so that it's ready to accept the stream
	ffmpeg.Start()
	raspivid.Start()
	raspivid.Wait()
	ffmpeg.Wait()
}

// Start a static file server for everything in the ./camera directory.
//
// This is a blocking operation.
func serveFiles() {
	fs := http.FileServer(http.Dir("./camera"))
	http.Handle("/", fs)

	log.Println("Server started on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}