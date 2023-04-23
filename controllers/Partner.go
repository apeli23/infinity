package controllers

import (
	"fmt"
	"net/http"

	"github.com/apeli23/infinity/database"
	"github.com/apeli23/infinity/models"
	"github.com/apeli23/infinity/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func MigratedToken(ctx *gin.Context) {
	type login struct {
		Username string `json:"AppKey" binding:"required"`
		Password string `json:"ApiSecret" binding:"required"`
	}
	loginInstance := login{}

	if err := ctx.ShouldBindJSON(&loginInstance); err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	partner, err := services.GetPartnerByEmail(loginInstance.Username)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	if !services.CheckPasswordHash(loginInstance.Password, partner.Secret) {
		logrus.Error("invalid login credentials")
		ctx.Status(http.StatusUnauthorized)
		return
	}

	token, err := services.GenerateToken(fmt.Sprintf("%d", partner.ID), "header enrichment")
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		"token": token,
	})

}

func CreatePartner(ctx *gin.Context) {

	partner := models.Partner{}

	if err := ctx.ShouldBindJSON(&partner); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	if err := services.CreatePartner(&partner); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusCreated, &partner)

}

func GetPartnerToken(ctx *gin.Context) {

	login := models.Login{}

	if err := ctx.ShouldBindJSON(&login); err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	partner, err := services.GetPartnerByEmail(login.Username)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}
	if !services.CheckPasswordHash(login.Password, partner.Secret) {
		logrus.Error("invalid login credentials")
		ctx.Status(http.StatusUnauthorized)
		return
	}

	token, err := services.GenerateToken(fmt.Sprintf("%d", partner.ID), "header enrichment")
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func GetAllPartners(ctx *gin.Context) {
	var partners []models.Partner
	if err := database.Db.Find(&partners).Error; err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": "Failed to fetch partners from database"})
		return
	}

	ctx.JSON(http.StatusOK, partners)
}