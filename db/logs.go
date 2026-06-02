package db

import (
	"database/sql"
	"time"
)

func (db *DB) CreateReviewLog(tx *sql.Tx, log *ReviewLog) (int64, error) {
	query := `
		INSERT INTO review_logs (card_id, reviewed_at, pre_repetition_count, pre_interval, pre_ease_factor, grade, post_repetition_count, post_interval, post_ease_factor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(query, log.CardID, log.ReviewedAt.Unix(), log.PreRepetitionCount, log.PreInterval, log.PreEaseFactor, log.Grade, log.PostRepetitionCount, log.PostInterval, log.PostEaseFactor)
	} else {
		res, err = db.Conn.Exec(query, log.CardID, log.ReviewedAt.Unix(), log.PreRepetitionCount, log.PreInterval, log.PreEaseFactor, log.Grade, log.PostRepetitionCount, log.PostInterval, log.PostEaseFactor)
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) DeleteLastReviewLog(tx *sql.Tx, cardID int64) (*ReviewLog, error) {
	var log ReviewLog
	var reviewedAtUnix int64
	querySelect := `
		SELECT id, card_id, reviewed_at, pre_repetition_count, pre_interval, pre_ease_factor, grade, post_repetition_count, post_interval, post_ease_factor
		FROM review_logs
		WHERE card_id = ?
		ORDER BY id DESC LIMIT 1
	`
	var err error
	if tx != nil {
		err = tx.QueryRow(querySelect, cardID).Scan(&log.ID, &log.CardID, &reviewedAtUnix, &log.PreRepetitionCount, &log.PreInterval, &log.PreEaseFactor, &log.Grade, &log.PostRepetitionCount, &log.PostInterval, &log.PostEaseFactor)
	} else {
		err = db.Conn.QueryRow(querySelect, cardID).Scan(&log.ID, &log.CardID, &reviewedAtUnix, &log.PreRepetitionCount, &log.PreInterval, &log.PreEaseFactor, &log.Grade, &log.PostRepetitionCount, &log.PostInterval, &log.PostEaseFactor)
	}
	if err != nil {
		return nil, err
	}
	log.ReviewedAt = time.Unix(reviewedAtUnix, 0)

	queryDelete := "DELETE FROM review_logs WHERE id = ?"
	if tx != nil {
		_, err = tx.Exec(queryDelete, log.ID)
	} else {
		_, err = db.Conn.Exec(queryDelete, log.ID)
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}

func (db *DB) GetDailyStreak() (int, error) {
	rows, err := db.Conn.Query("SELECT reviewed_at FROM review_logs ORDER BY reviewed_at DESC")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	reviewedDates := make(map[string]bool)
	for rows.Next() {
		var u int64
		if err := rows.Scan(&u); err != nil {
			return 0, err
		}
		t := time.Unix(u, 0).Local()
		reviewedDates[t.Format("2006-01-02")] = true
	}

	if len(reviewedDates) == 0 {
		return 0, nil
	}

	streak := 0
	checkDate := time.Now().Local()
	todayStr := checkDate.Format("2006-01-02")
	yesterdayStr := checkDate.AddDate(0, 0, -1).Format("2006-01-02")

	if !reviewedDates[todayStr] && !reviewedDates[yesterdayStr] {
		return 0, nil
	}

	if !reviewedDates[todayStr] {
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	for {
		dateStr := checkDate.Format("2006-01-02")
		if reviewedDates[dateStr] {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, nil
}

func (db *DB) GetRetentionAccuracy() (float64, error) {
	var total, success int
	err := db.Conn.QueryRow("SELECT COUNT(*), SUM(CASE WHEN grade >= 3 THEN 1 ELSE 0 END) FROM review_logs").Scan(&total, &success)
	if err != nil {
		return 0.0, nil
	}
	if total == 0 {
		return 0.0, nil
	}
	return (float64(success) / float64(total)) * 100.0, nil
}

func (db *DB) GetMasteryStats() (int, int, error) {
	var total, mastered int
	err := db.Conn.QueryRow("SELECT COUNT(*), SUM(CASE WHEN repetition_count >= 3 AND ease_factor >= 2.5 THEN 1 ELSE 0 END) FROM cards").Scan(&total, &mastered)
	if err != nil {
		return 0, 0, nil
	}
	return total, mastered, nil
}

func (db *DB) GetLast7DaysActivity() ([]int, error) {
	now := time.Now().Local()
	activity := make([]int, 7)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -6+i)
		startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location()).Unix()
		endOfDay := time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 999999999, day.Location()).Unix()

		var count int
		err := db.Conn.QueryRow("SELECT COUNT(*) FROM review_logs WHERE reviewed_at >= ? AND reviewed_at <= ?", startOfDay, endOfDay).Scan(&count)
		if err == nil {
			activity[i] = count
		}
	}
	return activity, nil
}
