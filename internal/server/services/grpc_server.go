package services

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/proto"
	"github.com/MPoline/alert_service_yp/internal/server/flags"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsServer реализует gRPC сервер для метрик
type MetricsServer struct {
	proto.UnimplementedMetricsServiceServer
	privKey *rsa.PrivateKey
	key     string
}

var (
	grpcServer    *grpc.Server
	metricsServer *MetricsServer
)

// InitGRPCServer инициализирует и запускает gRPC сервер
func InitGRPCServer(privKey *rsa.PrivateKey, key string) error {
	metricsServer = &MetricsServer{
		privKey: privKey,
		key:     key,
	}

	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(metricsServer.unaryInterceptor),
	)

	proto.RegisterMetricsServiceServer(grpcServer, metricsServer)

	lis, err := net.Listen("tcp", flags.FlagGRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", flags.FlagGRPCAddress, err)
	}

	go func() {
		zap.L().Info("Starting gRPC server", zap.String("address", flags.FlagGRPCAddress))
		if err := grpcServer.Serve(lis); err != nil {
			zap.L().Error("gRPC server failed", zap.Error(err))
		}
	}()

	return nil
}

// StopGRPCServer останавливает gRPC сервер
func StopGRPCServer() {
	if grpcServer != nil {
		zap.L().Info("Stopping gRPC server gracefully...")
		grpcServer.GracefulStop()
		zap.L().Info("gRPC server stopped")
	}
}

// unaryInterceptor перехватчик для обработки аутентификации и валидации
func (s *MetricsServer) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	zap.L().Debug("gRPC request",
		zap.String("method", info.FullMethod),
		zap.Any("request", req))

	return handler(ctx, req)
}

// UpdateMetrics обработчик для массового обновления метрик
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if req == nil || req.Metrics == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var errors []string
	var metrics []models.Metrics

	for _, metric := range req.Metrics {
		if m, err := s.processMetric(ctx, metric); err != nil {
			errors = append(errors, fmt.Sprintf("metric %s: %v", metric.Id, err))
			zap.L().Error("Failed to process metric",
				zap.String("metric_id", metric.Id),
				zap.Error(err))
		} else {
			metrics = append(metrics, m)
		}
	}

	if len(metrics) > 0 {
		if err := storage.MetricStorage.UpdateSliceOfMetrics(ctx, models.SliceMetrics{Metrics: metrics}); err != nil {
			errors = append(errors, fmt.Sprintf("batch update failed: %v", err))
			zap.L().Error("Failed to update metrics batch", zap.Error(err))
		}
	}

	if len(errors) > 0 {
		return &proto.UpdateMetricsResponse{
			Error: fmt.Sprintf("failed to process %d metrics: %v", len(errors), errors),
		}, nil
	}

	zap.L().Info("Metrics processed successfully via gRPC",
		zap.Int("metrics_count", len(req.Metrics)))

	return &proto.UpdateMetricsResponse{}, nil
}

// UpdateMetric обработчик для обновления одной метрики
func (s *MetricsServer) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	if req == nil || req.Metric == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	metric, err := s.processMetric(ctx, req.Metric)
	if err != nil {
		return &proto.UpdateMetricResponse{
			Error: fmt.Sprintf("failed to process metric %s: %v", req.Metric.Id, err),
		}, nil
	}

	if err := storage.MetricStorage.UpdateMetric(ctx, metric); err != nil {
		return &proto.UpdateMetricResponse{
			Error: fmt.Sprintf("failed to update metric %s: %v", req.Metric.Id, err),
		}, nil
	}

	zap.L().Info("Metric processed successfully via gRPC",
		zap.String("metric_id", req.Metric.Id))

	return &proto.UpdateMetricResponse{}, nil
}

// processMetric обрабатывает одну метрику и возвращает модель для storage
func (s *MetricsServer) processMetric(ctx context.Context, metric *proto.Metric) (models.Metrics, error) {
	if s.privKey != nil {
		if err := s.decryptMetricValue(metric); err != nil {
			return models.Metrics{}, fmt.Errorf("decryption failed: %w", err)
		}
	}

	if s.key != "" {
		if !s.verifyMetricHash(metric) {
			return models.Metrics{}, fmt.Errorf("signature verification failed")
		}
	}

	result := models.Metrics{
		ID:    metric.Id,
		MType: metric.Mtype,
	}

	switch metric.Mtype {
	case "counter":
		delta := metric.Delta
		result.Delta = &delta
	case "gauge":
		value := metric.Value
		result.Value = &value
	default:
		return models.Metrics{}, fmt.Errorf("unknown metric type: %s", metric.Mtype)
	}

	return result, nil
}

// decryptMetricValue расшифровывает значение метрики с использованием приватного ключа
func (s *MetricsServer) decryptMetricValue(metric *proto.Metric) error {
	if metric.Hash == "" || s.privKey == nil {
		return nil
	}

	encryptedData, err := hex.DecodeString(metric.Hash)
	if err != nil {
		return fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	decryptedData, err := crypto.DecryptLargeData(s.privKey, encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt metric data: %w", err)
	}

	switch metric.Mtype {
	case "counter":
		delta, err := strconv.ParseInt(string(decryptedData), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse decrypted counter value: %w", err)
		}
		metric.Delta = delta

	case "gauge":
		value, err := strconv.ParseFloat(string(decryptedData), 64)
		if err != nil {
			return fmt.Errorf("failed to parse decrypted gauge value: %w", err)
		}
		metric.Value = value

	default:
		return fmt.Errorf("unknown metric type for decryption: %s", metric.Mtype)
	}

	metric.Hash = ""

	zap.L().Debug("Metric decrypted successfully",
		zap.String("metric_id", metric.Id),
		zap.String("metric_type", metric.Mtype))

	return nil
}

// verifyMetricHash проверяет HMAC подпись метрики
func (s *MetricsServer) verifyMetricHash(metric *proto.Metric) bool {
	if metric.Hash == "" || s.key == "" {
		return true
	}

	originalHash := metric.Hash

	expectedHash := s.calculateMetricHash(metric)

	metric.Hash = originalHash

	if originalHash != expectedHash {
		zap.L().Warn("Metric hash verification failed",
			zap.String("metric_id", metric.Id),
			zap.String("received_hash", originalHash),
			zap.String("expected_hash", expectedHash))
		return false
	}

	zap.L().Debug("Metric hash verified successfully",
		zap.String("metric_id", metric.Id))
	return true
}

// calculateMetricHash вычисляет HMAC-SHA256 хеш для метрики
func (s *MetricsServer) calculateMetricHash(metric *proto.Metric) string {
	var data string
	switch metric.Mtype {
	case "counter":
		data = fmt.Sprintf("%s:counter:%d", metric.Id, metric.Delta)
	case "gauge":
		data = fmt.Sprintf("%s:gauge:%f", metric.Id, metric.Value)
	default:
		data = fmt.Sprintf("%s:%s", metric.Id, metric.Mtype)
	}

	h := hmac.New(sha256.New, []byte(s.key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
