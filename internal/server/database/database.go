package database

import (
	"database/sql"
	"errors"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

func OpenDBConnection() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", flags.FlagDatabaseDSN)
	if err != nil {
		zap.L().Error("Error opening database: ", zap.Error(err))
		return
	}
	zap.L().Info("Successful open to the database")
	return
}

func CloseDBConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		zap.L().Error("Error closing database: ", zap.Error(err))
	} else {
		zap.L().Info("The database connection was closed")
	}
}

func CreateMetricsTable(db *sql.DB) error {
	createQuery := ` CREATE TABLE IF NOT EXISTS metrics ( 
		id TEXT, 
		m_type TEXT CHECK(m_type IN ('gauge', 'counter')), 
		delta BIGINT, 
		value DOUBLE PRECISION,
		PRIMARY KEY (id, m_type)
	); `

	_, err := db.Exec(createQuery)
	if err != nil {
		zap.L().Error("Error create table metrics:", zap.Error(err))
		return err
	}
	zap.L().Info("Table metrics is exist")
	return nil
}

func CreateOrUpdateMetric(db *sql.DB, metric models.Metrics) error {

	query := ` INSERT INTO metrics 
		(id, m_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, m_type)
		DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value 
	; `

	_, err := db.Exec(query, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		zap.L().Error("SQL query execution error:", zap.Error(err))
		return err
	}
	zap.L().Info("Metric created/updated in metrics table")
	return nil
}

func GetAllMetricsFromDB(db *sql.DB) ([]models.Metrics, error) {
	var metrics []models.Metrics

	rows, err := db.Query(`SELECT id, m_type, delta, value FROM metrics`)
	if err != nil {
		zap.L().Error("Error getting all metrics:", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			zap.L().Error("Error in writing metric strings:", zap.Error(err))
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		zap.L().Error("Iteration error on results:", zap.Error(err))
		return nil, err
	}
	return metrics, nil
}

func GetOneMetric(db *sql.DB, id string, mType string) (models.Metrics, error) {
	var metric models.Metrics

	row := db.QueryRow(` SELECT id, m_type, delta, value FROM metrics WHERE id=$1 AND m_type=$2`,
		id, mType)

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err == sql.ErrNoRows {
		err = errors.New("MetricNotFound")
		zap.L().Info("Metric not found")
		return metric, err
	} else if err != nil {
		zap.L().Error("Error reading metrics", zap.Error(err))
		return metric, err
	}
	return metric, nil
}
