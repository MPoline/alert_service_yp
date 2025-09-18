package services

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/MPoline/alert_service_yp/internal/crypto"
	"github.com/MPoline/alert_service_yp/internal/models"
	"github.com/MPoline/alert_service_yp/internal/proto"
	"go.uber.org/zap"
)

type MetricProcessor struct {
	pubKey *rsa.PublicKey
	key    string
}

func NewMetricProcessor(pubKey *rsa.PublicKey, key string) *MetricProcessor {
	return &MetricProcessor{
		pubKey: pubKey,
		key:    key,
	}
}

// ProcessMetrics обрабатывает метрики: конвертирует, шифрует, вычисляет хеш
func (p *MetricProcessor) ProcessMetrics(metrics []models.Metrics) []*proto.Metric {
	protoMetrics := make([]*proto.Metric, 0, len(metrics))

	for _, m := range metrics {
		protoMetric, err := p.convertToProtoMetric(m)
		if err != nil {
			zap.L().Error("Failed to convert metric to proto",
				zap.String("metric_id", m.ID),
				zap.Error(err))
			continue
		}

		if p.pubKey != nil {
			if err := p.encryptMetricValue(protoMetric); err != nil {
				zap.L().Error("Failed to encrypt metric",
					zap.String("metric_id", m.ID),
					zap.Error(err))
				continue
			}
		}

		if p.key != "" {
			protoMetric.Hash = p.calculateMetricHash(protoMetric)
		}

		protoMetrics = append(protoMetrics, protoMetric)
	}

	return protoMetrics
}

func (p *MetricProcessor) convertToProtoMetric(m models.Metrics) (*proto.Metric, error) {
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

func (p *MetricProcessor) encryptMetricValue(metric *proto.Metric) error {
	var dataToEncrypt []byte

	switch metric.Mtype {
	case "counter":
		dataToEncrypt = []byte(strconv.FormatInt(metric.Delta, 10))
	case "gauge":
		dataToEncrypt = []byte(strconv.FormatFloat(metric.Value, 'f', -1, 64))
	default:
		return fmt.Errorf("unknown metric type for encryption: %s", metric.Mtype)
	}

	encryptedData, err := crypto.EncryptLargeData(p.pubKey, dataToEncrypt)
	if err != nil {
		return fmt.Errorf("failed to encrypt metric data: %w", err)
	}

	metric.Hash = hex.EncodeToString(encryptedData)
	return nil
}

func (p *MetricProcessor) calculateMetricHash(metric *proto.Metric) string {
	var data string
	switch metric.Mtype {
	case "counter":
		data = fmt.Sprintf("%s:counter:%d", metric.Id, metric.Delta)
	case "gauge":
		data = fmt.Sprintf("%s:gauge:%f", metric.Id, metric.Value)
	default:
		data = fmt.Sprintf("%s:%s", metric.Id, metric.Mtype)
	}

	h := hmac.New(sha256.New, []byte(p.key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}