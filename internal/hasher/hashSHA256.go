package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"

	"go.uber.org/zap"
)

// sha256Hasher реализует Hasher интерфейс используя HMAC-SHA256.
type sha256Hasher struct{}

// NewSHA265Hasher создает новый экземпляр SHA-256 HMAC хешера.
//
// Возвращает:
//   - *sha256Hasher: указатель на новый хешер
func NewSHA265Hasher() *sha256Hasher {
	return &sha256Hasher{}
}

// CalculateHash вычисляет HMAC-SHA256 хеш для данных с использованием ключа.
//
// Параметры:
//   - data: данные для хеширования
//   - key: секретный ключ
//
// Возвращает:
//   - []byte: вычисленный HMAC-SHA256 хеш
//   - error: ошибка если данные или ключ пустые, либо произошла ошибка записи данных
//
// Пример:
//
//	hasher := NewSHA265Hasher()
//	hash, err := hasher.CalculateHash([]byte("data"), []byte("secret"))
//	if err != nil {
//	    // обработка ошибки
//	}
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
