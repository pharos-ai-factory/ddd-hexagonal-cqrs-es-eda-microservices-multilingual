package postgres

import "context"

// Backlog reports pending work, failures and oldest pending age.
type Backlog struct {
	Pending          int64   `json:"pending"`
	OldestAgeSeconds float64 `json:"oldestAgeSeconds"`
	Failed           int64   `json:"failed"`
}

// Diagnostics reports owned backlog and worker recovery evidence.
type Diagnostics struct {
	Outbox   Backlog `json:"outbox"`
	Realtime Backlog `json:"realtime"`
}

func (db *ContextDatabase) Owner() string { return db.owner }
func (db *ContextDatabase) Diagnostics(ctx context.Context) (Diagnostics, error) {
	var result Diagnostics
	for index, tables := range [][2]string{{"dispatches", "outbox_events"}, {"realtime_dispatches", "realtime_publications"}} {
		backlog := &result.Outbox
		if index == 1 {
			backlog = &result.Realtime
		}
		err := db.pool.QueryRow(ctx, `SELECT count(*),COALESCE(EXTRACT(EPOCH FROM clock_timestamp()-min(o.created_at)),0)::float8,
   count(*) FILTER (WHERE d.last_error IS NOT NULL) FROM cafe.`+tables[0]+` d JOIN cafe.`+tables[1]+` o ON o.id=d.event_id
   WHERE d.completed_at IS NULL`).Scan(&backlog.Pending, &backlog.OldestAgeSeconds, &backlog.Failed)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}
