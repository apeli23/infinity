package models

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/apeli23/infinity/database"
)
// User: This structure represents a user of the application.
type User struct {
	//`json`:<string> specifies the JSON key to use for the <string> field when marshaling or unmarshaling JSON data
	//`gorm` tag provides metadata that is used by GORM ORM library to map the struct fields to dbase cols.
	ID        uint      `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	FirstName string    `json:"firstname" gorm:"column:firstname"`
	LastName  string    `json:"lastname" gorm:"column:lastname"`
	Email     string    `json:"email" gorm:"column:email"`
	Password  string    `json:"password" gorm:"column:password"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

//PartnerUsers: This structure represents the relationship between a partner and a user. It contains fields such as the ID of the user and the ID of the partner.
type PartnerUsers struct {
	ID        uint      `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	UserID    uint      `json:"user_id" gorm:"column:user_id"`
	PartnerID uint      `json:"partner_id" gorm:"column:partner_id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

//Plan: This structure represents a plan that a user can subscribe to. 
type Partner struct {
	ID          uint      `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	Name        string    `json:"name" gorm:"column:name"`
	Email       string    `json:"email" gorm:"column:email"`
	Secret      string    `json:"-" gorm:"column:secret"`
	PhoneNumber string    `json:"phone_number" gorm:"column:phone_number"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

//Plan: This structure represents a plan that a user can subscribe to. 
type Plan struct {
	ID        string    `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	Name      string    `json:"name" gorm:"column:name"`
	Amount    float64   `json:"amount" gorm:"column:cost"`
	Cycle     string    `json:"cycle" gorm:"column:frequency"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	PartnerID uint      `json:"partner" gorm:"column:partner_id"`
}

//Subscription: This structure represents a user's subscription to a plan.
type Subscription struct {
	ID                uint      `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	ExternalID        string    `json:"external_id" gorm:"column:external_id"`
	PlanID            string    `json:"plan" gorm:"column:plan_id"`
	MSISDN            string    `json:"msisdn" gorm:"column:msisdn"`
	Method            string    `json:"method" gorm:"column:method"`
	Status            string    `json:"status" gorm:"column:status"`
	StatusDescription string    `json:"description" gorm:"column:description"`
	Callback          string    `json:"callback" gorm:"column:callback"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"column:updated_at"`
}

//Transaction: This structure represents a transaction that occurs when a user's airtime is charged.
type Transaction struct {
	ID                uint      `json:"id" gorm:"column:id;primarykey;<-:false;autoIncrement;type:int"`
	ExternalID        string    `json:"external_id" gorm:"column:external_id"`
	SubscriptionID    uint      `json:"subscription" gorm:"column:subscription_id"`
	Status            string    `json:"status" gorm:"column:status"`
	StatusDescription string    `json:"description" gorm:"column:description"`
	Callback          string    `json:"callback" gorm:"column:callback"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"column:updated_at"`
	Amount            string    `json:"amount" gorm:"column:amount"`
}

//authentecation
type Login struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

//HeRequest: This structure represents a request to the Safaricom SDP to charge a user's airtime.
type HeRequest struct {
	ExternalID   string `json:"requestId" binding:"required"`
	Msisdn       string `json:"msisdn" binding:"required"`
	OfferCode    string `json:"offerCode" binding:"required"`
	CallBackUrl  string `json:"callBackUrl" binding:"required"`
	ChargeAmount string `json:"ChargeAmount" binding:"-"`
}

//HeResponseHeader: This structure represents the header of the response received from the Safaricom SDP after a request is sent.
type HeResponseHeader struct {
	RequestRefId    string      `json:"requestRefId" binding:"required"`
	ResponseCode    interface{} `json:"responseCode" binding:"required"`
	ResponseMessage string      `json:"responseMessage" binding:"required"`
	CustomerMessage string      `json:"customerMessage" binding:"required"`
	Timestamp       string      `json:"timestamp" binding:"required"`
}

//HeResponseBody: This structure represents the body of the response received from the Safaricom SDP after a request is sent.
type HeResponseBody struct {
	Status      string      `json:"status" binding:"required"`
	Description string      `json:"description" binding:"required"`
	StatusCode  string      `json:"statusCode"`
	Data        []dataItems `json:"data"`
}

//HeResponse struct: This struct defines the response of an API that returns two objects - HeResponseHeader and HeResponseBody. 
type HeResponse struct {
//The json tag is used to specify how the fields should be marshaled/unmarshaled from JSON.
	Header HeResponseHeader `json:"header" binding:"required"`
	Body   HeResponseBody   `json:"body" binding:"required"`
}

//This function is defined on a Subscription struct and it saves a subscription record to the database
func (sub *Subscription) SaveSubscription() (*Subscription, error) {
//It first tries to update an existing subscription record by matching the plan ID and MSISDN, 
// Debug() method on the Db object is used to enable debug logging for the query.
	result := database.Db.Debug().Table("subscriptions").Where("plan_id = ? AND msisdn = ?", sub.PlanID, sub.MSISDN).Updates(&sub)
	if result.Error != nil {
		return sub, result.Error
// if no record is found, it creates a new subscription record. 
//The RowsAffected property of the result object is used to check if any rows were affected by the update query.		
	} else if result.RowsAffected == 0 {
		if err := database.Db.Debug().Table("subscriptions").Save(&sub).Error; err != nil {
			logrus.Error(err)
			return sub, err
		}
	}

	return sub, nil
}

//This function is defined on a Transaction struct and it saves a transaction record to the database
func (trx *Transaction) SaveTransaction() (*Transaction, error) {
//It simply calls the Save() method on the transactions table and returns the result.
	if err := database.Db.Debug().Table("transactions").Save(&trx).Error; err != nil {
		logrus.Error(err)
		return trx, err
	}
	return trx, nil
}

//This struct represents a callback object that is received by an API
type Callback struct {
	RequestId    string       `json:"requestId"  binding:"required"`
	RequestParam requestParam `json:"requestParam"  binding:"required"`
}

// This struct represents the requestParam field of the Callback struct. It contains a Data field that is an array of dataItems.
type requestParam struct {
	Data []dataItems `json:"data"  binding:"required"`
}

//This struct represents an item in the Data array of the requestParam struct
type dataItems struct {
	Name  string      `json:"name"  binding:"required"`
	Value interface{} `json:"value"  binding:"required"`
}
