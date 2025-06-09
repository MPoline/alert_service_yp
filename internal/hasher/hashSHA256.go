package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	"go.uber.org/zap"
)

type sha256Hasher struct{}

func NewSHA265Hasher() *sha256Hasher {
	return &sha256Hasher{}
}

func (h *sha256Hasher) CalculateHash(data []byte, key []byte) (result []byte, err error) {
	if len(data) == 0 || len(key) == 0 {
		zap.L().Info("InputStringOrKeyIsEmpty")
		return nil, errors.New("InputStringOrKeyIsEmpty")
	}

	mac := hmac.New(sha256.New, key)

	_, err = mac.Write(data)
	if err != nil {
		zap.L().Error("Failed to write input data: ", zap.Error(err))
		return nil, err
	}

	result = mac.Sum(nil)
	return result, nil
}
