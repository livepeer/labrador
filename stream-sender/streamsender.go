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
	http := flag.String("http", "localhost:5000", "http address to run the web server on")
	interval := flag.Duration("interval", 2*time.Hour, "interval to blast streams into the networks")
	streamTester := flag.String("server", "localhost:3001", "http address the stream-tester server is running on")
	broadcaster := flag.String("broadcaster", "localhost", "ip of the broadcaster")

	flag.Parse()

	// Create a channel to receive OS signals
	c := make(chan os.Signal)
	// Relay os.Interrupt to our channel (os.Interrupt = CTRL+C)
	// Ignore other incoming signals
	signal.Notify(c, os.Interrupt)

	cfg := &stream.Config{
		Host:            *broadcaster,
		Rtmp:            1935,
		Media:           8935,
		FileName:        "official_test_source_2s_keys_24pfs.mp4",
		Repeat:          1,
		Simultaneous:    2,
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
	fmt.Println("sleeping for 30 seconds before sending streams")
	fmt.Printf("\n \n %v\n", fmt.Sprintf(strings.Repeat("*", 60)))

	streamErr := make(chan error, 1)
	go func() {
		time.Sleep(30 * time.Second)

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
