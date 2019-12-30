package store

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/livepeer/stream-sender/models"
	_ "github.com/mattn/go-sqlite3" // blank import
)

const dbPath = "/tmp/streamsender/streamsender.db"

// DB is an initialized DB driver with prepared statements
type DB struct {
	dbh *sql.DB

	insertStats *sql.Stmt
	selectStats *sql.Stmt
	allStats    *sql.Stmt
}

var schema = `
	CREATE TABLE IF NOT EXISTS stats (
		baseManifestID STRING PRIMARY_KEY,
		rtmpStreams INTEGER,
		mediaStreams INTEGER,
		totalSegments INTEGER,
		sentSegments INTEGER,
		downloadedSegments INTEGER,
		totalDownloadSegments INTEGER,
		failedToDownloadSegments INTEGER,
		profilesNum INTEGER,
		retries INTEGER,
		successRate STRING,
		connectionLost INTEGER,
		finished BOOLEAN,
		sourceLatencies BLOB,
		transcodedLatencies BLOB,
		gaps INTEGER,
		startTime int64
	)
`
var version = 1

// InitDB initializes a DB instance
func InitDB() (*DB, error) {
	//Make sure dbPath is present
	if _, err := os.Stat("/tmp/streamsender/"); os.IsNotExist(err) {
		if err = os.MkdirAll("/tmp/streamsender/", 0755); err != nil {
			return nil, fmt.Errorf("error making /tmp/streamsender/: %v", err)
		}
	}
	d := &DB{}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("error opening sql DB: %v", err)
	}
	db.SetMaxOpenConns(1)
	d.dbh = db
	schemaBuf := new(bytes.Buffer)
	tmpl := template.Must(template.New("schema").Parse(schema))
	tmpl.Execute(schemaBuf, version)
	_, err = db.Exec(schemaBuf.String())
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("error executing schema: %v", err)
	}

	stmt, err := db.Prepare(`
	INSERT OR REPLACE INTO stats(baseManifestID, rtmpStreams, mediaStreams, totalSegments, sentSegments, downloadedSegments, totalDownloadSegments, failedToDownloadSegments, profilesNum, retries, successRate, connectionLost, finished, sourceLatencies, transcodedLatencies, gaps, startTime)
	VALUES(:baseManifestID, :rtmpStreams, :mediaStreams, :totalSegments, :sentSegments, :downloadedSegments, :totalDownloadSegments, :failedToDownloadSegments, :profilesNum, :retries, :successRate, :connectionLost, :finished, :sourceLatencies, :transcodedLatencies, :gaps, :startTime)
	`)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("error preparing insertStats statement: %v", err)
	}
	d.insertStats = stmt

	stmt, err = db.Prepare("SELECT * FROM stats WHERE baseManifestID = ?")
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("error preparing selectStats statement: %v", err)
	}
	d.selectStats = stmt

	stmt, err = db.Prepare("SELECT * FROM stats ORDER BY startTime DESC")
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("error preparing allStats statement: %v", err)
	}
	d.allStats = stmt
	return d, nil
}

// Close the DB connection
func (db *DB) Close() error {
	if db.insertStats != nil {
		db.insertStats.Close()
	}
	if db.selectStats != nil {
		db.selectStats.Close()
	}
	if db.allStats != nil {
		db.allStats.Close()
	}
	return db.dbh.Close()
}

// InsertStats inserts streaming statistics for a manifestID
func (db *DB) InsertStats(manifestID string, stats *models.Stats) error {
	startTime := stats.StartTime.UnixNano()

	sourceLats, err := json.Marshal(stats.SourceLatencies)
	if err != nil {
		return err
	}

	transcodedLats, err := json.Marshal(stats.TranscodedLatencies)
	if err != nil {
		return err
	}

	_, err = db.insertStats.Exec(
		sql.Named("baseManifestID", manifestID),
		sql.Named("rtmpStreams", stats.RTMPstreams),
		sql.Named("mediaStreams", stats.MediaStreams),
		sql.Named("totalSegments", stats.TotalSegmentsToSend),
		sql.Named("sentSegments", stats.SentSegments),
		sql.Named("downloadedSegments", stats.DownloadedSegments),
		sql.Named("totalDownloadSegments", stats.ShouldHaveDownloadedSegments),
		sql.Named("failedToDownloadSegments", stats.FailedToDownloadSegments),
		sql.Named("profilesNum", stats.ProfilesNum),
		sql.Named("retries", stats.Retries),
		sql.Named("successRate", fmt.Sprintf("%f", stats.SuccessRate)),
		sql.Named("connectionLost", stats.ConnectionLost),
		sql.Named("finished", stats.Finished),
		sql.Named("sourceLatencies", sourceLats),
		sql.Named("transcodedLatencies", transcodedLats),
		sql.Named("gaps", stats.Gaps),
		sql.Named("startTime", startTime),
	)
	return err
}

// SelectStats for a stream by manifest ID
func (db *DB) SelectStats(manifestID string) (*models.Stats, error) {
	var (
		baseManifestID               string
		rtmpStreams                  int
		mediaStreams                 int
		totalSegmentsToSend          int
		sentSegments                 int
		downloadedSegments           int
		shouldHaveDownloadedSegments int
		failedToDownloadSegments     int
		profilesNum                  int
		retries                      int
		successRate                  string
		connectionLost               int
		finished                     bool
		sourceLatencies              []byte
		transcodedLatencies          []byte
		gaps                         int
		startTime                    int64
	)
	if err := db.selectStats.QueryRow(manifestID).Scan(
		&baseManifestID,
		&rtmpStreams,
		&mediaStreams,
		&totalSegmentsToSend,
		&sentSegments,
		&downloadedSegments,
		&shouldHaveDownloadedSegments,
		&failedToDownloadSegments,
		&profilesNum,
		&retries,
		&successRate,
		&connectionLost,
		&finished,
		&sourceLatencies,
		&transcodedLatencies,
		&gaps,
		&startTime,
	); err != nil {
		return nil, err
	}

	var sourceL models.Latencies
	var transcodedL models.Latencies
	err := json.Unmarshal(sourceLatencies, &sourceL)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(transcodedLatencies, &transcodedL)
	if err != nil {
		return nil, err
	}

	success, err := strconv.ParseFloat(successRate, 64)
	if err != nil {
		return nil, err
	}

	return &models.Stats{
		RTMPstreams:                  rtmpStreams,
		MediaStreams:                 mediaStreams,
		TotalSegmentsToSend:          totalSegmentsToSend,
		SentSegments:                 sentSegments,
		DownloadedSegments:           downloadedSegments,
		ShouldHaveDownloadedSegments: shouldHaveDownloadedSegments,
		FailedToDownloadSegments:     failedToDownloadSegments,
		ProfilesNum:                  profilesNum,
		Retries:                      retries,
		SuccessRate:                  success,
		ConnectionLost:               connectionLost,
		Finished:                     finished,
		SourceLatencies:              sourceL,
		TranscodedLatencies:          transcodedL,
		Gaps:                         gaps,
		StartTime:                    time.Unix(0, startTime),
	}, nil
}

// AllStats return stats for all streams
func (db *DB) AllStats() (map[string]*models.Stats, error) {
	all := make(map[string]*models.Stats)

	rows, err := db.allStats.Query()
	defer rows.Close()
	if err != nil {
		return all, nil
	}

	for rows.Next() {
		var (
			baseManifestID               string
			rtmpStreams                  int
			mediaStreams                 int
			totalSegmentsToSend          int
			sentSegments                 int
			downloadedSegments           int
			shouldHaveDownloadedSegments int
			failedToDownloadSegments     int
			profilesNum                  int
			retries                      int
			successRate                  string
			connectionLost               int
			finished                     bool
			sourceLatencies              []byte
			transcodedLatencies          []byte
			gaps                         int
			startTime                    int64
		)

		if err := rows.Scan(
			&baseManifestID,
			&rtmpStreams,
			&mediaStreams,
			&totalSegmentsToSend,
			&sentSegments,
			&downloadedSegments,
			&shouldHaveDownloadedSegments,
			&failedToDownloadSegments,
			&profilesNum,
			&retries,
			&successRate,
			&connectionLost,
			&finished,
			&sourceLatencies,
			&transcodedLatencies,
			&gaps,
			&startTime,
		); err != nil {
			fmt.Println(err)
			continue
		}

		var sourceL models.Latencies
		var transcodedL models.Latencies
		err := json.Unmarshal(sourceLatencies, &sourceL)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = json.Unmarshal(transcodedLatencies, &transcodedL)
		if err != nil {
			fmt.Println(err)
			continue
		}

		success, err := strconv.ParseFloat(successRate, 64)
		if err != nil {
			fmt.Println(err)
			continue
		}

		all[baseManifestID] = &models.Stats{
			RTMPstreams:                  rtmpStreams,
			MediaStreams:                 mediaStreams,
			TotalSegmentsToSend:          totalSegmentsToSend,
			SentSegments:                 sentSegments,
			DownloadedSegments:           downloadedSegments,
			ShouldHaveDownloadedSegments: shouldHaveDownloadedSegments,
			FailedToDownloadSegments:     failedToDownloadSegments,
			ProfilesNum:                  profilesNum,
			Retries:                      retries,
			SuccessRate:                  success,
			ConnectionLost:               connectionLost,
			Finished:                     finished,
			SourceLatencies:              sourceL,
			TranscodedLatencies:          transcodedL,
			Gaps:                         gaps,
			StartTime:                    time.Unix(0, startTime),
		}
	}
	return all, nil
}
