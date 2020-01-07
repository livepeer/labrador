package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/livepeer/stream-sender/store"
	"github.com/livepeer/stream-sender/stream"
)

// HTTPServer an HTTP server instance for streamsender
type HTTPServer struct {
	address  string
	db       *store.DB
	streamer *stream.Streamer
}

// NewHTTPServer returns a new HTTPServer instance
func NewHTTPServer(address string, db *store.DB, streamer *stream.Streamer) *HTTPServer {
	return &HTTPServer{
		address,
		db,
		streamer,
	}
}

// StartServer starts the HTTP server
func (s *HTTPServer) StartServer() error {
	mux := s.setupHandlers()
	server := &http.Server{
		Addr:    s.address,
		Handler: mux,
	}

	return server.ListenAndServe()
}

func (s *HTTPServer) setupHandlers() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/stats/all", s.allStreams)
	mux.HandleFunc("/stats/select", s.selectStream)
	mux.HandleFunc("/stream/start", s.startStream)
	mux.HandleFunc("/config/update", s.updateConfig)
	mux.HandleFunc("/config", s.getConfig)
	return mux
}

func (s *HTTPServer) allStreams(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.db.AllStats()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(stats)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	w.Write(b)
}

func (s *HTTPServer) selectStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var statsReq struct {
		BaseManifestID string `json:"base_manifest_id"`
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.Unmarshal(body, &statsReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	stats, err := s.db.SelectStats(statsReq.BaseManifestID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	b, err := json.Marshal(stats)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *HTTPServer) startStream(w http.ResponseWriter, r *http.Request) {

	// Config preflight request
	s.preflight(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var cfg stream.Config
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.Unmarshal(body, &cfg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	cfg.DoNotClearStats = false

	mid, err := s.streamer.SendStreamRequest(&cfg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	res, err := json.Marshal(
		map[string]string{
			"success":          "true",
			"base_manifest_id": mid,
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	w.Write(res)
}

func (s *HTTPServer) updateConfig(w http.ResponseWriter, r *http.Request) {

	// Config preflight request
	s.preflight(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var cfg stream.Config
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Error("err reading body", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if err := json.Unmarshal(body, &cfg); err != nil {
		glog.Error("err unmarshaling json", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	cfg.DoNotClearStats = false

	s.streamer.SetConfig(&cfg)

	w.Write([]byte{})
}

func (s *HTTPServer) getConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cfg := s.streamer.GetConfig()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(cfg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(b)
}

func (s *HTTPServer) preflight(w http.ResponseWriter, r *http.Request) {
	// PREFLIGHT SETUP
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
