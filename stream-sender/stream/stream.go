package stream

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/livepeer/stream-sender/models"
)

const httpTimeout = 8 * time.Second

// Streamer streams into a stream-tester server on a periodic interval and saves the resulting statistics into storage
type Streamer struct {
	cfg    *Config
	server string
	client *http.Client
	ticker *time.Ticker
	quit   chan interface{}
	stats  models.StatsStore
	mu     sync.Mutex
}

// Config to start streaming
type Config struct {
	Host            string `json:"host"`         // Host name of broadcaster to stream to
	Rtmp            int    `json:"rtmp"`         // Port number to stream RTMP stream to
	Media           int    `json:"media"`        // Port number to download media from
	FileName        string `json:"file_name"`    // Path to file to stream (should exists in local filesystem of streamer)
	Repeat          int    `json:"repeat"`       // How many times to repeat streaming
	Simultaneous    int    `json:"simultaneous"` // How many simultaneous streams stream into broadcaster
	ProfilesNum     int    `json:"profiles_num"` // How many transcoding profiles broadcaster configured with
	DoNotClearStats bool   `json:"do_not_clear_stats"`
	MeasureLatency  bool   `json:"measure_latency"`
}

type sendStreamResponse struct {
	Success        bool   `json:"success"`
	BaseManifestID string `json:"base_manifest_id"`
}

// NewStreamer returns a new Streamer instance
func NewStreamer(cfg *Config, server string, interval time.Duration, stats models.StatsStore) *Streamer {
	return &Streamer{
		cfg:    cfg,
		server: "http://" + server,
		client: &http.Client{
			Timeout: httpTimeout,
		},
		ticker: time.NewTicker(interval),
		quit:   make(chan interface{}),
		stats:  stats,
	}
}

// Start streaming
func (s *Streamer) Start() error {
	mid, err := s.SendStreamRequest(s.GetConfig())
	if err != nil {
		return err
	}
	glog.Infof(">> Started stream with base manifest ID %v", mid)

	for {
		select {
		case <-s.ticker.C:
			mid, err := s.SendStreamRequest(s.GetConfig())
			if err != nil {
				glog.Error(err)
			}
			glog.Infof(">> Started stream with base manifest ID %v", mid)
		case <-s.quit:
			return nil
		}
	}
}

// Stop all running streams
func (s *Streamer) Stop() error {
	timeout := 8 * time.Second

	client := &http.Client{
		Timeout: timeout,
	}

	res, err := client.Get(s.server + "/stop")
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("unable to stop streams: %v", res.Status)
	}

	s.ticker.Stop()
	close(s.quit)
	return nil
}

// SendStreamRequest sends a request to start streams
func (s *Streamer) SendStreamRequest(cfg *Config) (string, error) {
	cfg.MeasureLatency = true
	in, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.server+"/start_streams", bytes.NewBuffer(in))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("unable to make http request: %v", res.Status)
	}

	var resJSON sendStreamResponse
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response body: %v", err)
	}
	if err := json.Unmarshal(b, &resJSON); err != nil {
		return "", fmt.Errorf("unable to unmarshal response body: %v", err)
	}

	if !resJSON.Success {
		return "", fmt.Errorf("server failed to start streams")
	}

	go s.pollAndFlushStats(resJSON.BaseManifestID)

	return resJSON.BaseManifestID, nil
}

func (s *Streamer) SetConfig(cfg *Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg
}

func (s *Streamer) GetConfig() *Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

// pollAndFlushStats waits for a stream to finish and then writes the statistics to the database
// It is upon the caller to implement concurrency
func (s *Streamer) pollAndFlushStats(manifestID string) {
	var stats models.Stats
	req, err := http.NewRequest("GET", s.server+"/stats?latencies&base_manifest_id="+url.QueryEscape(manifestID), nil)
	if err != nil {
		glog.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	for !stats.Finished {
		// wait 30 seconds to make sure server has manifests available
		time.Sleep(30 * time.Second)

		res, err := s.client.Do(req)
		if err != nil {
			glog.Error(err)
			return
		}

		if res.StatusCode != 200 {
			glog.Errorf("unable to make http request: %v", res.Status)
			return
		}

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			glog.Errorf("unable to read response body: %v", err)
			return
		}

		if err := json.Unmarshal(b, &stats); err != nil {
			glog.Errorf("unable to unmarshal response body: %v", err)
			return
		}

		if err := s.stats.InsertStats(manifestID, &stats); err != nil {
			glog.Errorf("unable to insert stats into DB: %v", err)
		}
	}
}
