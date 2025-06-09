package storage

import (
	"context"
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

	err = database.CreateMetricsTable(context.Background(), dbConn)
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

func (s DBStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	metrics, err := database.GetAllMetricsFromDB(ctx, s.dbConn)
	if err != nil {
		zap.L().Fatal("Error get all metric from table: ", zap.Error(err))
		return nil, err
	}
	return metrics, nil
}

func (s DBStorage) GetMetric(ctx context.Context, metricType string, metricName string) (models.Metrics, error) {
	select {
	case <-ctx.Done():
		return models.Metrics{}, ctx.Err()
	default:
	}

	metric, err := database.GetOneMetric(ctx, s.dbConn, metricName, metricType)
	if err != nil {
		zap.L().Error("Error get metric from table: ", zap.Error(err))
		return metric, err
	}
	return metric, nil
}

func (s DBStorage) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := metric.IsValid()
	if err != nil {
		zap.L().Info("Error in Metric Parametrs")
		return err
	}

	if metric.MType == "counter" {
		m, err := s.GetMetric(ctx, metric.MType, metric.ID)

		if err == nil {
			*metric.Delta += *m.Delta
		} else if err.Error() != "MetricNotFound" {
			zap.L().Error("Error create/update metric from table: ", zap.Error(err))
			return err
		}
	}

	err = database.CreateOrUpdateMetric(ctx, s.dbConn, metric)
	if err != nil {
		zap.L().Error("Error create/update metric from table: ", zap.Error(err))
		return err
	}

	return nil
}

func (s DBStorage) UpdateSliceOfMetrics(ctx context.Context, sliceMitrics models.SliceMetrics) error {
	for _, metric := range sliceMitrics.Metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if ok, err := metric.IsValid(); !ok {
			zap.L().Info("Error in Metric Parametrs: ", zap.Error(err))
			return err
		}
	}

	for _, metric := range sliceMitrics.Metrics {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if metric.MType == "counter" {
			m, err := s.GetMetric(ctx, metric.MType, metric.ID)

			if err == nil {
				*metric.Delta += *m.Delta
			} else if err.Error() != "MetricNotFound" {
				zap.L().Error("Error create/update metric from table: ", zap.Error(err))
				return err
			}
		}
	}

	err := database.CreateOrUpdateSliceOfMetrics(ctx, s.dbConn, sliceMitrics)
	if err != nil {
		zap.L().Error("Error create/update metric from table: ", zap.Error(err))
		return err
	}

	zap.L().Info("All metrics updated successfully")
	return nil
}
