package services

import (
	"net/http"

	"github.com/MPoline/alert_service_yp/internal/storage"
	"github.com/gin-gonic/gin"
)

func CheckDBConnection(c *gin.Context) {
	db, err := storage.DBInit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error opening database"})
		return
	}

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error checking database connection"})
		return
	}
	c.String(http.StatusOK, "Successful connection to the database")
}
