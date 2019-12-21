package models

import "time"

// Stats represents global test statistics
type Stats struct {
	RTMPstreams                  int             `json:"rtmp_streams"`  // number of RTMP streams
	MediaStreams                 int             `json:"media_streams"` // number of media streams
	TotalSegmentsToSend          int             `json:"total_segments_to_send"`
	SentSegments                 int             `json:"sent_segments"`
	DownloadedSegments           int             `json:"downloaded_segments"`
	ShouldHaveDownloadedSegments int             `json:"should_have_downloaded_segments"`
	FailedToDownloadSegments     int             `json:"failed_to_download_segments"`
	Retries                      int             `json:"retries"`
	SuccessRate                  float64         `json:"success_rate"` // DownloadedSegments/profilesNum*SentSegments
	ConnectionLost               int             `json:"connection_lost"`
	Finished                     bool            `json:"finished"`
	RawSourceLatencies           []time.Duration `json:"raw_source_latencies"`
	RawTranscodedLatencies       []time.Duration `json:"raw_transcoded_latencies"`
	Gaps                         int             `json:"gaps"`
	StartTime                    time.Time       `json:"start_time"`
}
