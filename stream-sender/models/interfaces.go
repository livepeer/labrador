package models

// StatsStore represent the interface for storage of stream statistics
type StatsStore interface {
	InsertStats(manifestID string, stats *Stats) error
	SelectStats(manifestID string) (*Stats, error)
	AllStats() (map[string]*Stats, error)
}
