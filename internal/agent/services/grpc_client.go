package services

import (
	"context"
	"crypto/rsa"
	"fmt"
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
	client          proto.MetricsServiceClient
	conn            *grpc.ClientConn
	metricProcessor *MetricProcessor
}

func NewGRPCClient() (*GRPCClient, error) {
	address := flags.FlagGRPCAddress
	if address == "" {
		address = flags.FlagRunAddr
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsServiceClient(conn)

	var pubKey *rsa.PublicKey
	if flags.FlagCryptoKey != "" {
		pubKey, err = crypto.LoadPublicKey(flags.FlagCryptoKey)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
	}

	metricProcessor := NewMetricProcessor(pubKey, flags.FlagKey)

	grpcClient := &GRPCClient{
		client:          client,
		conn:            conn,
		metricProcessor: metricProcessor,
	}

	zap.L().Info("gRPC client initialized successfully",
		zap.String("address", address))
	return grpcClient, nil
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			zap.L().Error("Failed to close gRPC connection", zap.Error(err))
		} else {
			zap.L().Info("GPRC connection closed successfully")
		}
	}
}

// SendMetrics отправляет метрики на сервер через gRPC
func (c *GRPCClient) SendMetrics(memStorage *storage.MemStorage, metrics []models.Metrics, localIP string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	protoMetrics := c.metricProcessor.ProcessMetrics(metrics)

	if len(protoMetrics) == 0 {
		zap.L().Warn("No valid metrics to send via gRPC")
		return
	}

	req := &proto.UpdateMetricsRequest{Metrics: protoMetrics}
	resp, err := c.client.UpdateMetrics(ctx, req)
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

func (c *GRPCClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx, &proto.PingRequest{})
	if err != nil {
		return fmt.Errorf("gRPC health check failed: %w", err)
	}

	return nil
}
