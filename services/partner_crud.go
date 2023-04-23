package services

import (
	"github.com/sirupsen/logrus"
	"github.com/apeli23/infinity/database"
	"github.com/apeli23/infinity/models"
)

func CreatePartner(partner *models.Partner) (err error) {

	password := GeneratePassword()
	logrus.Info(password, partner.Email)
	partner.Secret, _ = HashPassword(password)

	if err = database.Db.Debug().Table("partners").Save(partner).Error; err != nil {
		logrus.Error(err)
		return
	}

	return
}

func GetPartnerByEmail(email string) (partner models.Partner, err error) {

	if err = database.Db.Debug().Table("partners").Where("email = ?", email).First(&partner).Error; err != nil {
		return
	}
	return

}

func PartnerCheckPlanAccess(userId, planId string) (plan models.Plan) {
	// plan := models.Plan{}
	if err := database.Db.Debug().Table("plans").Where("id = ? AND partner_id = ?", planId, userId).First(&plan).Error; err != nil {
		logrus.Error(err)
		return
	}
	return

}