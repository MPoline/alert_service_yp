// Package services содержит бизнес-логику обработки HTTP-запросов сервера метрик.
package services

import (
	"net/http"

	"github.com/MPoline/alert_service_yp/internal/server/database"
	"github.com/gin-gonic/gin"
)

// CheckDBConnection обрабатывает запрос проверки соединения с базой данных.
//
// Эндпоинт: GET /ping
//
// Логика работы:
//  1. Устанавливает соединение с БД
//  2. Проверяет доступность БД методом Ping()
//  3. Возвращает результат проверки
//
// Возможные ответы:
//  - 200 OK: соединение успешно установлено
//    Тело ответа: "Successful connection to the database"
//  - 500 Internal Server Error: ошибка соединения
//    Тело ответа: {"Error": "Error opening database"} 
//                или {"Error": "Error checking database connection"}
//
// Пример использования:
//  router := gin.Default()
//  router.GET("/ping", services.CheckDBConnection)
func CheckDBConnection(c *gin.Context) {
	db, err := database.OpenDBConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error opening database"})
		return
	}
	defer database.CloseDBConnection(db)

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error checking database connection"})
		return
	}
	c.String(http.StatusOK, "Successful connection to the database")
}
