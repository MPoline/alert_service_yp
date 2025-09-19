package services

import (
	"fmt"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/storage"
	"go.uber.org/zap"
)

// MetricClient интерфейс для клиентов отправки метрик
type MetricClient interface {
	SendMetrics(memStorage *storage.MemStorage, metrics []models.Metrics, localIP string)
	HealthCheck() error
	Close()
}

// ClientManager управляет клиентами для отправки метрик
type ClientManager struct {
	client MetricClient
}

func NewClientManager() (*ClientManager, error) {
	if flags.FlagGRPC {
		zap.L().Info("Initializing gRPC client",
			zap.String("address", flags.FlagGRPCAddress))

		grpcClient, err := NewGRPCClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC client: %w", err)
		}
		return &ClientManager{client: grpcClient}, nil
	} else {
		zap.L().Info("Using HTTP protocol",
			zap.String("address", flags.FlagRunAddr))

		httpClient := NewHTTPClient()
		return &ClientManager{client: httpClient}, nil
	}
}

func (m *ClientManager) SendMetrics(memStorage *storage.MemStorage, metrics []models.Metrics, localIP string) {
	if m.client != nil {
		m.client.SendMetrics(memStorage, metrics, localIP)
	} else {
		zap.L().Error("Client not initialized")
	}
}

func (m *ClientManager) Close() {
	if m.client != nil {
		m.client.Close()
	}
}

func (m *ClientManager) HealthCheck() error {
	if m.client != nil {
		return m.client.HealthCheck()
	}
	return fmt.Errorf("client not initialized")
}