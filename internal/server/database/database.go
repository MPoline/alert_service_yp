// Package database предоставляет функциональность для работы с PostgreSQL:
// - Установка соединения с БД
// - Создание и управление структурой таблиц
// - CRUD операции с метриками
// - Транзакционная обработка данных
//
// Пакет использует драйвер github.com/lib/pq и поддерживает:
// - Обработку ошибок PostgreSQL
// - Логирование операций через zap
package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

// SQL запрос для вставки или обновления метрики
var createOrUpdateQuery = `INSERT INTO metrics 
	(id, m_type, delta, value)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id, m_type)
	DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value;`

// OpenDBConnection устанавливает соединение с PostgreSQL.
//
// Использует DSN из flags.FlagDatabaseDSN.
// Возвращает:
//   - *sql.DB: объект соединения с БД
//   - error: ошибка соединения
//
// Пример:
//
//	db, err := database.OpenDBConnection()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer database.CloseDBConnection(db)
func OpenDBConnection() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", flags.FlagDatabaseDSN)
	if err != nil {
		zap.L().Error("Error opening database: ", zap.Error(err))
		return
	}
	zap.L().Info("Successful open to the database")
	return
}

// CloseDBConnection закрывает соединение с БД.
// Логирует ошибку закрытия при ее возникновении.
func CloseDBConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		zap.L().Error("Error closing database: ", zap.Error(err))
	} else {
		zap.L().Info("The database connection was closed")
	}
}

// CreateMetricsTable создает таблицу metrics если она не существует.
// Параметры:
//   - ctx context.Context: контекст выполнения
//   - db *sql.DB: соединение с БД
//
// Возвращает:
//   - error: ошибка выполнения запроса
func CreateMetricsTable(ctx context.Context, db *sql.DB) error {
	createQuery := `CREATE TABLE IF NOT EXISTS metrics ( 
		id TEXT, 
		m_type TEXT CHECK(m_type IN ('gauge', 'counter')), 
		delta BIGINT, 
		value DOUBLE PRECISION,
		PRIMARY KEY (id, m_type)
	);`

	_, err := db.ExecContext(ctx, createQuery)
	if err != nil {
		handlePGError(err)
		return err
	}
	zap.L().Info("Table metrics is exist")
	return nil
}

// CreateOrUpdateMetric создает или обновляет метрику в БД.
// Параметры:
//   - ctx context.Context: контекст выполнения
//   - db *sql.DB: соединение с БД
//   - metric models.Metrics: метрика для сохранения
//
// Возвращает:
//   - error: ошибка выполнения операции
func CreateOrUpdateMetric(ctx context.Context, db *sql.DB, metric models.Metrics) error {
	_, err := db.ExecContext(ctx, createOrUpdateQuery, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		handlePGError(err)
		return err
	}
	zap.L().Info("Metric created/updated in metrics table")
	return nil
}

// CreateOrUpdateSliceOfMetrics создает или обновляет несколько метрик в транзакции.
// Параметры:
//   - ctx context.Context: контекст выполнения
//   - db *sql.DB: соединение с БД
//   - metrics models.SliceMetrics: список метрик
//
// Возвращает:
//   - error: ошибка выполнения операции
func CreateOrUpdateSliceOfMetrics(ctx context.Context, db *sql.DB, metrics models.SliceMetrics) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, metric := range metrics.Metrics {
		_, err := tx.ExecContext(ctx, createOrUpdateQuery, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			handlePGError(err)
			tx.Rollback()
			return err
		}
	}
	zap.L().Info("Metric created/updated within transaction")
	return tx.Commit()
}

// GetAllMetricsFromDB возвращает все метрики из БД.
// Параметры:
//   - ctx context.Context: контекст выполнения
//   - db *sql.DB: соединение с БД
//
// Возвращает:
//   - []models.Metrics: список метрик
//   - error: ошибка выполнения запроса
func GetAllMetricsFromDB(ctx context.Context, db *sql.DB) ([]models.Metrics, error) {
	var metrics []models.Metrics

	rows, err := db.QueryContext(ctx, `SELECT id, m_type, delta, value FROM metrics`)
	if err != nil {
		handlePGError(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var metric models.Metrics
		err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			handlePGError(err)
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		handlePGError(err)
		return nil, err
	}
	return metrics, nil
}

// GetOneMetric возвращает одну метрику по ID и типу.
// Параметры:
//   - ctx context.Context: контекст выполнения
//   - db *sql.DB: соединение с БД
//   - id string: идентификатор метрики
//   - mType string: тип метрики ('gauge' или 'counter')
//
// Возвращает:
//   - models.Metrics: найденная метрика
//   - error: ошибка выполнения запроса (MetricNotFound если метрика не найдена)
func GetOneMetric(ctx context.Context, db *sql.DB, id string, mType string) (models.Metrics, error) {
	var metric models.Metrics

	row := db.QueryRowContext(ctx, `SELECT id, m_type, delta, value FROM metrics WHERE id=$1 AND m_type=$2`,
		id, mType)

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err == sql.ErrNoRows {
		err = errors.New("MetricNotFound")
		zap.L().Info("Metric not found")
		return metric, err
	} else if err != nil {
		handlePGError(err)
		return metric, err
	}
	return metric, nil
}

// handlePGError обрабатывает и логирует ошибки PostgreSQL.
// Различает различные типы ошибок БД:
//   - Нарушение уникальности
//   - Нарушение внешнего ключа
//   - Нарушение проверочного ограничения
//   - Нарушение NOT NULL
func handlePGError(err error) {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			zap.L().Warn("PostgreSQL unique constraint violation",
				zap.String("constraint_name", pgErr.ConstraintName),
				zap.String("detail", pgErr.Detail))
		case pgerrcode.ForeignKeyViolation:
			zap.L().Warn("PostgreSQL foreign key constraint violation",
				zap.String("constraint_name", pgErr.ConstraintName),
				zap.String("detail", pgErr.Detail))
		case pgerrcode.CheckViolation:
			zap.L().Warn("PostgreSQL check constraint violation",
				zap.String("constraint_name", pgErr.ConstraintName),
				zap.String("detail", pgErr.Detail))
		case pgerrcode.NotNullViolation:
			zap.L().Warn("PostgreSQL not-null constraint violation",
				zap.String("column_name", pgErr.ColumnName),
				zap.String("table_name", pgErr.TableName))
		default:
			zap.L().Error("Unhandled PostgreSQL error:", zap.Error(pgErr))
		}
	} else {
		zap.L().Error("Unknown error:", zap.Error(err))
	}
}
