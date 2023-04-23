package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/apeli23/infinity/database"
	"github.com/apeli23/infinity/models"
	"github.com/apeli23/infinity/services"
	"github.com/apeli23/infinity/utils"
)

func ActivationDeactivationNotification(ctx *gin.Context) {

	notification := models.Callback{}
	if err := ctx.ShouldBindJSON(&notification); err != nil {
		logrus.Error(err)
		return
	}
	services.ActDeactProcess(notification)

}

func ChargeNotification(ctx *gin.Context) {
	notification := models.Callback{}
	transaction := models.Transaction{}
	offercode := ""

	for _, data := range notification.RequestParam.Data {

		switch data.Name {
		case "ClientTransactionId":
			transaction.ExternalID = data.Value.(string)
		case "OfferCode":
			offercode = data.Value.(string)
		case "Reason":
			transaction.Status = data.Value.(string)
		}

	}

	if transaction.Status == "Successful" {
		transaction.StatusDescription = "Subscriber charged"
	} else {
		transaction.StatusDescription = transaction.Status
	}

	// NOTE: Better to use partner_id and external_id.
	// But plan_id is okay given that it is a 1 to 1 representation of customer.
	if err := database.Db.Debug().Table("transactions_v").Where("external_id = ? AND plan_id = ?", transaction.ExternalID, offercode).First(&transaction).Error; err != nil {
		logrus.Error(err)
		return
	}
	err := database.Db.Debug().Table("transactions").Where("id = ?", transaction.ExternalID, transaction.ID).Updates(&transaction).Error
	if err != nil {
		logrus.Error(err)
		return
	}
	payload, _ := json.Marshal(notification)
	utils.Request(string(payload), map[string][]string{
		"Content-Type": {"application/json"},
	}, transaction.Callback, "POST")

}

func ActivateSubscriber(ctx *gin.Context) {
	activation := models.HeRequest{}
	partnerId := ctx.GetString("user_id")

	if err := ctx.ShouldBindJSON(&activation); err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	plan := services.PartnerCheckPlanAccess(partnerId, activation.OfferCode)
	if plan.ID == "" {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	response, err := services.SendActivation(&activation, "USSD")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusAccepted, response)

}

func WebActivateSubscriber(ctx *gin.Context) {
	activation := models.HeRequest{}
	partnerId := ctx.GetString("user_id")

	if err := ctx.ShouldBindJSON(&activation); err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	plan := services.PartnerCheckPlanAccess(partnerId, activation.OfferCode)
	if plan.ID == "" {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	response, err := services.WebActivation(&activation, "WEB")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusAccepted, response)

}

func DeActivateSubscriber(ctx *gin.Context) {
	deactivation := models.HeRequest{}
	partnerId := ctx.GetString("user_id")

	if err := ctx.ShouldBindJSON(&deactivation); err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	plan := services.PartnerCheckPlanAccess(partnerId, deactivation.OfferCode)
	if plan.ID == "" {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	response, err := services.SendDeActivation(&deactivation, "USSD")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusAccepted, response)

}

func ChargeSubscriber(ctx *gin.Context) {
	charging := models.HeRequest{}
	partnerId := ctx.GetString("user_id")

	if err := ctx.ShouldBindJSON(&charging); err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	plan := services.PartnerCheckPlanAccess(partnerId, charging.OfferCode)
	if plan.ID == "" {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	response, err := services.SendCharging(&charging)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.AbortWithStatusJSON(http.StatusAccepted, response)

}