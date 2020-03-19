package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/livepeer/stream-sender/server"
	"github.com/livepeer/stream-sender/store"
	"github.com/livepeer/stream-sender/stream"
)

func main() {
	http := flag.String("http", "localhost:5000", "http address to run the web server on (default: localhost:5000)")
	interval := flag.Duration("interval", 1*time.Hour, "interval to blast streams into the networks (default: 1h)")
	streamTester := flag.String("server", "localhost:3001", "http address the stream-tester server is running on (default: 3001)")
	broadcaster := flag.String("broadcaster", "localhost", "ip of the broadcaster (default: localhost)")
	rtmpPort := flag.Int("rtmpPort", 1935, "broadcaster rtmp port (default: 1935)")
	mediaPort := flag.Int("mediaPort", 8935, "http port for the broadcaster (default 8935)")
	fileName := flag.String("file", "bbb_sunflower_1080p_30fps_normal_t02.mp4", "video file to transcode (file must be present in the root directory of stream-tester)")
	simultaneous := flag.Int("simultaneous", 2, "number of concurrent streams to run (default: 2)")
	dbPath := flag.String("dbPath", "/tmp/streamsender", "path to DB")
	flag.Parse()

	// Create a channel to receive OS signals
	c := make(chan os.Signal)
	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)

	cfg := &stream.Config{
		Host:            *broadcaster,
		Rtmp:            *rtmpPort,
		Media:           *mediaPort,
		FileName:        *fileName,
		Repeat:          1,
		Simultaneous:    *simultaneous,
		ProfilesNum:     3,
		DoNotClearStats: false,
	}

	db, err := store.InitDB()
	if err != nil {
		glog.Error(err)
		return
	}
	defer db.Close()

	streamer := stream.NewStreamer(cfg, *streamTester, *interval, db)
	defer func() {
		if err := streamer.Stop(); err != nil {
			glog.Error(err)
		}
	}()

	httpServerErr := make(chan error, 1)
	go func() {
		srv := server.NewHTTPServer(*http, db, streamer)
		if err := srv.StartServer(); err != nil {
			httpServerErr <- err
		}
	}()

	fmt.Printf(" %v \n \n", fmt.Sprintf(strings.Repeat("*", 60)))
	fmt.Printf("Stream sender started, blasting new streams every %v hours to broadcaster at %v\n", *interval, *broadcaster)
	fmt.Printf("sending %v copies of %v repeated %v times\n", cfg.Simultaneous, cfg.FileName, cfg.Repeat)
	fmt.Println()
	fmt.Println("sleeping for 60 seconds before sending streams")
	fmt.Printf("\n \n %v\n", fmt.Sprintf(strings.Repeat("*", 60)))

	streamErr := make(chan error, 1)
	go func() {
		time.Sleep(60 * time.Second)

		if err := streamer.Start(); err != nil {
			streamErr <- err
		}
	}()

	select {
	case <-c:
		glog.Info("stopping stream sender...")
		return
	case err := <-streamErr:
		glog.Error(err)
		return
	case err := <-httpServerErr:
		glog.Error(err)
		return
	}
}
