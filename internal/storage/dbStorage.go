package storage

import (
	"database/sql"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/database"
	"go.uber.org/zap"
)

type DBStorage struct {
	dbConn *sql.DB
}

func NewDBStorage() *DBStorage {
	dbConn, err := database.OpenDBConnection()
	if err != nil {
		zap.L().Fatal("Error opening database: ", zap.Error(err))
	}

	err = database.CreateMetricsTable(dbConn)
	if err != nil {
		zap.L().Fatal("Error create table metric: ", zap.Error(err))
	}

	return &DBStorage{
		dbConn: dbConn,
	}
}

func (s DBStorage) Close() {
	database.CloseDBConnection(s.dbConn)
}

func (s DBStorage) GetAllMetrics() ([]models.Metrics, error) {
	metrics, err := database.GetAllMetricsFromDB(s.dbConn)
	if err != nil {
		zap.L().Fatal("Error get all metric from table: ", zap.Error(err))
		return nil, err
	}
	return metrics, nil
}

func (s DBStorage) GetMetric(metricType string, metricName string) (models.Metrics, error) {
	metric, err := database.GetOneMetric(s.dbConn, metricName, metricType)
	if err != nil {
		zap.L().Error("Error get metric from table: ", zap.Error(err))
		return metric, err
	}
	return metric, nil
}

func (s DBStorage) UpdateMetric(metric models.Metrics) error {
	_, err := metric.IsValid()
	if err != nil {
		zap.L().Info("Error in Metric Parametrs")
		return err
	}

	if metric.MType == "counter" {
		m, err := s.GetMetric(metric.MType, metric.ID)

		if err == nil {
			*metric.Delta += *m.Delta
		} else if err.Error() != "MetricNotFound" {
			zap.L().Error("Error create/update metric from table: ", zap.Error(err))
			return err
		}
	}

	err = database.CreateOrUpdateMetric(s.dbConn, metric)
	if err != nil {
		zap.L().Error("Error create/update metric from table: ", zap.Error(err))
		return err
	}

	return nil
}

func (s DBStorage) UpdateSliceOfMetrics(sliceMitrics models.SliceMetrics) error {
	for _, metric := range sliceMitrics.Metrics {
		if ok, err := metric.IsValid(); !ok {
			zap.L().Info("Error in Metric Parametrs: ", zap.Error(err))
			return err
		}
	}

	for _, metric := range sliceMitrics.Metrics {
		if metric.MType == "counter" {
			m, err := s.GetMetric(metric.MType, metric.ID)

			if err == nil {
				*metric.Delta += *m.Delta
			} else if err.Error() != "MetricNotFound" {
				zap.L().Error("Error create/update metric from table: ", zap.Error(err))
				return err
			}
		}
	}

	err := database.CreateOrUpdateSliceOfMetrics(s.dbConn, sliceMitrics)
	if err != nil {
		zap.L().Error("Error create/update metric from table: ", zap.Error(err))
		return err
	}

	zap.L().Info("All metrics updated successfully")
	return nil
}
