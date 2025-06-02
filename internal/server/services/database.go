package services

import (
	"net/http"

	"github.com/MPoline/alert_service_yp/internal/server/database"
	"github.com/gin-gonic/gin"
)

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
