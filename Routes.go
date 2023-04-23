package main

import (
	"github.com/gin-gonic/gin"
	"github.com/apeli23/infinity/controllers"
)

var basePath = "/api/v2"

// Routes Function to route mapping
var Routes = map[string]map[string]gin.HandlerFunc{
	basePath + "/partners": {
		"GET": controllers.GetAllPartners,
	},
	basePath + "/partner/add": {
		"POST": controllers.CreatePartner,
	},
	"/public/v2/partner/token": {
		"POST": controllers.GetPartnerToken,
	},
	basePath + "/ussd/activation": {
		"POST": controllers.ActivateSubscriber,
	},
	basePath + "/web/activation": {
		"POST": controllers.WebActivateSubscriber,
	},
	basePath + "/ussd/deactivation": {
		"POST": controllers.DeActivateSubscriber,
	},
	basePath + "/ussd/charge": {
		"POST": controllers.ChargeSubscriber,
	},
	"/public/v2/notification/subscription": {
		"POST": controllers.ActivationDeactivationNotification,
	},
	"/public/v2/notification/charge": {
		"POST": controllers.ChargeNotification,
	},
	//NOTE: Old versions migrated from previous version
	"/api/v1/he/activation": {
		"POST": controllers.ActivateSubscriber,
	},
	"/api/v1/he/deactivation": {
		"POST": controllers.DeActivateSubscriber,
	},
	"/api/v1/he/charge": {
		"POST": controllers.ChargeSubscriber,
	},
	"/public/token/:service": {
		"POST": controllers.MigratedToken,
	},
}