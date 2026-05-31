package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/krishnassh/picostatus/internal/config"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SyncChecks(cfgChecks []config.Check) error {
	for _, c := range cfgChecks {
		_, err := r.db.Exec(`
			INSERT INTO checks (name, target, interval, timeout)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				target   = excluded.target,
				interval = excluded.interval,
				timeout  = excluded.timeout
		`, c.Name, c.URL, c.Interval.Nanoseconds(), c.Timeout.Nanoseconds())
		if err != nil {
			return fmt.Errorf("sync check %q: %w", c.Name, err)
		}
	}
	return nil
}

func (r *Repository) GetChecks() ([]Check, error) {
	rows, err := r.db.Query(`SELECT id, name, target, interval, timeout FROM checks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []Check
	for rows.Next() {
		var c Check
		var intervalNs, timeoutNs int64
		if err := rows.Scan(&c.ID, &c.Name, &c.Target, &intervalNs, &timeoutNs); err != nil {
			return nil, err
		}
		c.Interval = time.Duration(intervalNs)
		c.Timeout = time.Duration(timeoutNs)
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func (r *Repository) InsertResult(checkID int64, status Status, latencyMS int64, checkErr string, retainN int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO results (check_id, status, latency_ms, error, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, checkID, string(status), latencyMS, checkErr, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("insert result: %w", err)
	}

	_, err = tx.Exec(`
		DELETE FROM results
		WHERE check_id = ?
		  AND id NOT IN (
			SELECT id FROM results
			WHERE check_id = ?
			ORDER BY created_at DESC, id DESC
			LIMIT ?
		  )
	`, checkID, checkID, retainN)
	if err != nil {
		return fmt.Errorf("prune results: %w", err)
	}

	now := time.Now().Unix()
	uptime1h, err := calcUptime(tx, checkID, now-3600, now)
	if err != nil {
		return fmt.Errorf("calc uptime 1h: %w", err)
	}
	uptime24h, err := calcUptime(tx, checkID, now-86400, now)
	if err != nil {
		return fmt.Errorf("calc uptime 24h: %w", err)
	}
	uptime7d, err := calcUptime(tx, checkID, now-604800, now)
	if err != nil {
		return fmt.Errorf("calc uptime 7d: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO uptime (check_id, uptime_1h, uptime_24h, uptime_7d)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(check_id) DO UPDATE SET
			uptime_1h  = excluded.uptime_1h,
			uptime_24h = excluded.uptime_24h,
			uptime_7d  = excluded.uptime_7d
	`, checkID, uptime1h, uptime24h, uptime7d)
	if err != nil {
		return fmt.Errorf("upsert uptime: %w", err)
	}

	return tx.Commit()
}

func calcUptime(tx *sql.Tx, checkID int64, from, to int64) (float64, error) {
	var total, up int64
	err := tx.QueryRow(`
		SELECT
			COUNT(*),
			SUM(CASE WHEN status = 'up' THEN 1 ELSE 0 END)
		FROM results
		WHERE check_id = ? AND created_at >= ? AND created_at <= ?
	`, checkID, from, to).Scan(&total, &up)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return float64(up) / float64(total), nil
}

func (r *Repository) GetLatestResults() ([]Result, error) {
	rows, err := r.db.Query(`
		SELECT id, check_id, status, latency_ms, error, created_at
		FROM results
		WHERE id IN (
			SELECT id FROM results
			GROUP BY check_id
			HAVING created_at = MAX(created_at)
		)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func (r *Repository) GetCheckHistory(checkID int64, limit int) ([]Result, error) {
	rows, err := r.db.Query(`
		SELECT id, check_id, status, latency_ms, error, created_at
		FROM results
		WHERE check_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, checkID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanResults(rows)
}

func (r *Repository) GetUptime(checkID int64) (Uptime, error) {
	var u Uptime
	err := r.db.QueryRow(`
		SELECT check_id, uptime_1h, uptime_24h, uptime_7d
		FROM uptime WHERE check_id = ?
	`, checkID).Scan(&u.CheckID, &u.Uptime1h, &u.Uptime24h, &u.Uptime7d)
	if err == sql.ErrNoRows {
		return Uptime{CheckID: checkID}, nil // no data yet
	}
	return u, err
}

func (r *Repository) GetAllUptimes() (map[int64]Uptime, error) {
	rows, err := r.db.Query(`SELECT check_id, uptime_1h, uptime_24h, uptime_7d FROM uptime`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]Uptime)
	for rows.Next() {
		var u Uptime
		if err := rows.Scan(&u.CheckID, &u.Uptime1h, &u.Uptime24h, &u.Uptime7d); err != nil {
			return nil, err
		}
		out[u.CheckID] = u
	}
	return out, rows.Err()
}

func scanResults(rows *sql.Rows) ([]Result, error) {
	var results []Result
	for rows.Next() {
		var res Result
		var createdAtUnix int64
		if err := rows.Scan(&res.ID, &res.CheckID, &res.Status, &res.LatencyMS, &res.Error, &createdAtUnix); err != nil {
			return nil, err
		}
		res.CreatedAt = time.Unix(createdAtUnix, 0)
		results = append(results, res)
	}
	return results, rows.Err()
}
