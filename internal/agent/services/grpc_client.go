package services

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/proto"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	client proto.MetricsServiceClient
	conn   *grpc.ClientConn
	pubKey *rsa.PublicKey
}

var grpcClient *GRPCClient

// InitGRPCClient инициализирует gRPC клиент для отправки метрик
func InitGRPCClient() error {
	address := flags.FlagGRPCAddress
	if address == "" {
		address = flags.FlagRunAddr
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsServiceClient(conn)

	var pubKey *rsa.PublicKey
	if flags.FlagCryptoKey != "" {
		pubKey, err = crypto.LoadPublicKey(flags.FlagCryptoKey)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to load public key: %w", err)
		}
	}

	grpcClient = &GRPCClient{
		client: client,
		conn:   conn,
		pubKey: pubKey,
	}

	zap.L().Info("gRPC client initialized successfully",
		zap.String("address", address))
	return nil
}

func CloseGRPCClient() {
	if grpcClient != nil && grpcClient.conn != nil {
		if err := grpcClient.conn.Close(); err != nil {
			zap.L().Error("Failed to close gRPC connection", zap.Error(err))
		} else {
			zap.L().Info("GPRC connection closed successfully")
		}
	}
}

// SendMetricsGRPC отправляет метрики на сервер через gRPC
func SendMetricsGRPC(memStorage *storage.MemStorage, metrics []models.Metrics, localIP string) {
	if grpcClient == nil {
		zap.L().Error("gRPC client not initialized")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	protoMetrics := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		protoMetric, err := convertToProtoMetric(m)
		if err != nil {
			zap.L().Error("Failed to convert metric to proto",
				zap.String("metric_id", m.ID),
				zap.Error(err))
			continue
		}

		if grpcClient.pubKey != nil {
			if err := encryptMetricValue(protoMetric, grpcClient.pubKey); err != nil {
				zap.L().Error("Failed to encrypt metric",
					zap.String("metric_id", m.ID),
					zap.Error(err))
				continue
			}
		}

		if flags.FlagKey != "" {
			protoMetric.Hash = calculateMetricHash(protoMetric, flags.FlagKey)
		}

		protoMetrics = append(protoMetrics, protoMetric)
	}

	if len(protoMetrics) == 0 {
		zap.L().Warn("No valid metrics to send via gRPC")
		return
	}

	req := &proto.UpdateMetricsRequest{Metrics: protoMetrics}
	resp, err := grpcClient.client.UpdateMetrics(ctx, req)
	if err != nil {
		zap.L().Error("Failed to send metrics via gRPC",
			zap.Error(err),
			zap.String("agent_ip", localIP),
			zap.Int("metrics_count", len(protoMetrics)))
		return
	}

	if resp.Error != "" {
		zap.L().Error("gRPC server returned error",
			zap.String("error", resp.Error),
			zap.String("agent_ip", localIP))
		return
	}

	zap.L().Info("Metrics sent successfully via gRPC",
		zap.Int("metrics_count", len(protoMetrics)),
		zap.String("agent_ip", localIP))
}

// convertToProtoMetric конвертирует внутреннюю модель метрики в protobuf
func convertToProtoMetric(m models.Metrics) (*proto.Metric, error) {
	protoMetric := &proto.Metric{
		Id:    m.ID,
		Mtype: m.MType,
	}

	switch m.MType {
	case "counter":
		if m.Delta == nil {
			return nil, fmt.Errorf("counter metric %s has no delta value", m.ID)
		}
		protoMetric.Delta = *m.Delta

	case "gauge":
		if m.Value == nil {
			return nil, fmt.Errorf("gauge metric %s has no value", m.ID)
		}
		protoMetric.Value = *m.Value

	default:
		return nil, fmt.Errorf("unknown metric type: %s", m.MType)
	}

	return protoMetric, nil
}

// encryptMetricValue шифрует значение метрики с использованием публичного ключа
func encryptMetricValue(metric *proto.Metric, pubKey *rsa.PublicKey) error {
	var dataToEncrypt []byte

	switch metric.Mtype {
	case "counter":
		dataToEncrypt = []byte(strconv.FormatInt(metric.Delta, 10))
	case "gauge":
		dataToEncrypt = []byte(strconv.FormatFloat(metric.Value, 'f', -1, 64))
	default:
		return fmt.Errorf("unknown metric type for encryption: %s", metric.Mtype)
	}

	encryptedData, err := crypto.EncryptLargeData(pubKey, dataToEncrypt)
	if err != nil {
		return fmt.Errorf("failed to encrypt metric data: %w", err)
	}

	metric.Hash = hex.EncodeToString(encryptedData)

	return nil
}

// calculateMetricHash вычисляет HMAC-SHA256 хеш для метрики
func calculateMetricHash(metric *proto.Metric, key string) string {
	var data string
	switch metric.Mtype {
	case "counter":
		data = fmt.Sprintf("%s:counter:%d", metric.Id, metric.Delta)
	case "gauge":
		data = fmt.Sprintf("%s:gauge:%f", metric.Id, metric.Value)
	default:
		data = fmt.Sprintf("%s:%s", metric.Id, metric.Mtype)
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// HealthCheck проверяет доступность gRPC сервера
func HealthCheck() error {
	if grpcClient == nil {
		return fmt.Errorf("gRPC client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := grpcClient.client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{})
	if err != nil {
		return fmt.Errorf("gRPC health check failed: %w", err)
	}

	return nil
}
