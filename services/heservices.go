package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/apeli23/infinity/database"
	"github.com/apeli23/infinity/models"
	"github.com/apeli23/infinity/utils"
)

//this function performs a login and returns an access token. 
func HeLoginToken() (token string, err error) {

	//check if the token is already cached in memory, and if so, it returns the cached token
	if val, ok := utils.CacheInstance.Get("HE_TOKEN"); ok {
		return val.(string), nil
	}
	// Otherwise, it makes an HTTP request to the authentication endpoint with the provided credentials...
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", os.Getenv("HE_USERNAME"), os.Getenv("HE_PASSWORD"))))
	//...and parses the response JSON to extract the access token. 
	res, err := utils.Request("", map[string][]string{
		"Content-Type":  {"application/x-www-form-urlencoded"},
		"Authorization": {fmt.Sprintf("Basic %s", auth)},
	}, os.Getenv("HE_AUTH_URL"), "POST")

	if err != nil {
		logrus.Error(err)
		return
	}

	response := map[string]interface{}{}

	if err = json.Unmarshal([]byte(res), &response); err != nil {
		logrus.Error(err)
		return
	}
	// cache the access token in memory with a TTL of 50 minutes and returns it
	token = response["access_token"].(string)
	utils.CacheInstance.Set("HE_TOKEN", response["access_token"], 50*time.Minute)
	return

}

//function to retrieve a token for authenticating against a third-party service
//retruns a token
func GetSdpToken() (token string, err error) {

	//if the token is already chached return cached token
	if val, ok := utils.CacheInstance.Get("SDP_TOKEN"); ok {
		return val.(string), nil
	}
	//If the token is not cached, construct a payload that includes the SDP username and password.
	payload := fmt.Sprintf(`{"username":"%s","password":"%s"}`, os.Getenv("SDP_USERNAME"), os.Getenv("SDP_PASSWORD"))

	//send HTTP POST request to the SDP authentication URL, including the payload headers
	//response in JSON

	res, err := utils.Request(payload, map[string][]string{
		"Content-Type":     {`application/json`},
		"Accept":           {`application/json`},
		"X-Requested-With": {"XMLHttpRequest"},
	}, os.Getenv("SDP_AUTH_URL"), "POST")

	if err != nil {
		logrus.Error(err)
		return
	}

	response := map[string]string{}

	if err = json.Unmarshal([]byte(res), &response); err != nil {
		logrus.Error(err)
		return
	}
// the token from the JSON response and caches it for future use.
	token = response["token"]

	utils.CacheInstance.Set("SDP_TOKEN", token, 50*time.Minute)
	return

}

//Below function sends activation requests
func SendActivation(activation *models.HeRequest, channel string) (heResponse models.HeResponse, err error) {
	//construct a url based on `HE_BASE_URL`
	url := fmt.Sprintf("%s/api/v1/activate", os.Getenv("HE_BASE_URL"))
	//construct a payload in JSON format using data from the activation parameter and other environment variables.
	payoad := fmt.Sprintf(`{
		"msisdn": "%s",
		"offerCode": "%s",
		"CpId": "%s",
		"callBackUrl": "%s"
	}`, activation.Msisdn, activation.OfferCode, os.Getenv("CPID"), os.Getenv("ACT_DEACT_NOTIFICATION"))

	// the `BuildHeaders` function builds a map of HTTP headers that will be included in the API request
	headers, err := BuildHeaders(activation.ExternalID)

	if err != nil {
		logrus.Error(err)
		return
	}

	response, err := utils.Request(payoad, headers, url, "POST")
	if err != nil {
		logrus.Error(err)
		err = errors.New(response)
		return
	}
	heResponse, err = HeResponseProcessing(activation, response, channel)
	return

}

// function builds and returns a map of HTTP headers that need to be included in API requests
//It takes requesId as an argument, which is used to set X-Correlation-Conversation-ID and X-MessageID headers.
func BuildHeaders(requesId string) (map[string][]string, error) {
	//get required tokens
	//If HeLoginToken or GetSdpToken returns an error, the function returns nil and the error.
	heToken, err := HeLoginToken()
	if err != nil {
		return nil, err
	}
	sdpToken, err := GetSdpToken()
	if err != nil {
		return nil, err
	}

	headers := map[string][]string{
		"Authorization":                 {fmt.Sprintf("Bearer %s", heToken)},// Bearer token for HE login.
		"X-api-auth-token":              {fmt.Sprintf("Bearer %s", sdpToken)},//Bearer token for SDP authentication.
		"X-Api-Key":                     {os.Getenv("X_API_KEY")},// API key
		"Accept-Encoding":               {"application/json"},//: Indicates the encoding of the response that the client can understand.
		"Accept-Language":               {"EN"},// Language preferences of the client.
		"Content-Type":                  {"application/json"},//Type of data being sent in the request payload.
		"X-App":                         {"ussd"},//Application name.
		"X-Correlation-Conversation-ID": {requesId},// Conversation ID to correlate requests and responses.
		"X-MessageID":                   {requesId},// Unique identifier of the request.
		"X-Source-Division":             {"DIT"},//Division name.
		"X-Source-CountryCode":          {"KE"},//Country code of the source.
		"X-Source-Operator":             {"Safaricom"},//Operator of the source.
		"X-Source-System":               {"web-portal"},// System name of the source.
		"X-Source-Timestamp":            {fmt.Sprintf("%d", time.Now().Unix())},//Timestamp of the request.
		"X-Version":                     {"1.0.0"},// API version
	}
	return headers, nil
}
func SendDeActivation(activation *models.HeRequest, channel string) (heResponse models.HeResponse, err error) {
	url := fmt.Sprintf("%s/api/v1/deactivate", os.Getenv("HE_BASE_URL"))
	payoad := fmt.Sprintf(`{
		"msisdn": "%s",
		"offerCode": "%s",
		"CpId": "%s",
		"callBackUrl": "%s"
	}`, activation.Msisdn, activation.OfferCode, os.Getenv("CPID"), os.Getenv("ACT_DEACT_NOTIFICATION"))

	headers, err := BuildHeaders(activation.ExternalID)

	if err != nil {
		return
	}

	response, err := utils.Request(payoad, headers, url, "POST")
	if err != nil {
		logrus.Error(err)
		return
	}
	heResponse, err = HeResponseProcessing(activation, response, channel)
	return

}

//Below function responsible for sending a charging request to the HE API.
func SendCharging(chargeRequest *models.HeRequest) (heResponse models.HeResponse, err error) {
	//construct  the URL to the HE API endpoint for charging requests using the HE_BASE_URL environment variable.
	url := fmt.Sprintf("%s/api/v1/charge", os.Getenv("HE_BASE_URL"))
	// heResponse := models.HeResponse{}
	subscription := models.Subscription{}
	//create payload
	payoad := fmt.Sprintf(`{
		"msisdn": "%s",
		"offerCode": "%s",
		"CpId": "%s",
		"ChargeAmount": "%s",
		"callBackUrl": "%s"
	}`, chargeRequest.Msisdn, chargeRequest.OfferCode, os.Getenv("CPID"), chargeRequest.ChargeAmount, os.Getenv("CHARGE_CALLBACK"))

	//build headers for the request using the BuildHeaders function, passing the ExternalID value from the chargeRequest parameter.
	headers, err := BuildHeaders(chargeRequest.ExternalID)

	if err != nil {
		return
	}
	//the charging request to the HE API using the utils.Request function, passing the payload, headers, URL, and the HTTP method (POST).
	response, err := utils.Request(payoad, headers, url, "POST")
	if err != nil {
		return
	}
	//unmarshal  the response from the HE API into the heResponse variable.
	err = json.Unmarshal([]byte(response), &heResponse)
	if err != nil {
		logrus.Error(err)
		return
	}

	//fetch  the subscription information for the msisdn and offerCode from the database.
	if err = database.Db.Debug().Table("subscriptions").Where("msisdn = ? AND plan_id = ?",
		chargeRequest.Msisdn, chargeRequest.OfferCode).First(&subscription).Error; err != nil {
		logrus.Error(err)
		return
	}

	//create transaction object
	transaction := models.Transaction{
		ExternalID:        chargeRequest.ExternalID,
		SubscriptionID:    subscription.ID,
		Status:            heResponse.Body.Status,
		StatusDescription: heResponse.Body.Description,
		Amount:            chargeRequest.ChargeAmount,
		Callback:          chargeRequest.CallBackUrl,
	}
	//save the Transaction object to the database using the SaveTransaction method.
	_, err = transaction.SaveTransaction()
	return

}


// below function response from the HE API after an activation or deactivation request has been made.
func HeResponseProcessing(activation *models.HeRequest, response, channel string) (heResponse models.HeResponse, err error) {
	//unmarshal  the response into a models.HeResponse struct
	err = json.Unmarshal([]byte(response), &heResponse)
	if err != nil {
		logrus.Error(err)
		return
	}
// a models.Subscription struct using data from the activation parameter and the heResponse parameter.
	sub := models.Subscription{
		ExternalID:        activation.ExternalID,
		PlanID:            activation.OfferCode,
		MSISDN:            activation.Msisdn,
		Callback:          activation.CallBackUrl,
		Method:            channel,
		Status:            heResponse.Body.Status,
		StatusDescription: heResponse.Body.Description,
	}

	//save  subscription to the database using the SaveSubscription method of the Subscription struct.
	_, err = sub.SaveSubscription()
	return

}

//below function that performs a web activation request to a third-party service.
func WebActivation(activation *models.HeRequest, channel string) (heResponse models.HeResponse, err error) {
// build the URL to send the activation request 
	url := fmt.Sprintf("%s/api/v1/wapActivate", os.Getenv("HE_BASE_URL"))
//build the payload to send in the request.
	payoad := fmt.Sprintf(`{
		"msisdn": "%s",
		"offerCode": "%s",
		"CpId": "%s",
		"callBackUrl": "%s"
	}`, activation.Msisdn, activation.OfferCode, os.Getenv("CPID"), os.Getenv("ACT_DEACT_NOTIFICATION"))
// build the request headers using the BuildHeaders function
	headers, err := BuildHeaders(activation.ExternalID)

	if err != nil {
		logrus.Error(err)
		return
	}
// send the request using the utils.Request function.
	response, err := utils.Request(payoad, headers, url, "POST")
	if err != nil {
		logrus.Error(err)
		err = errors.New(response)
		return
	}
	// successful requests processes the response using the HeResponseProcessing function and returns a models.HeResponse struct.
	heResponse, err = HeResponseProcessing(activation, response, channel)
	return

}

//Below function takes in a callback notification received from an external system and updates the subscription status in the local database accordingly.
func ActDeactProcess(notification models.Callback) {
// initialize empty Subscription struct
	subscription := models.Subscription{}

//extract relevant information from the notification
	for _, data := range notification.RequestParam.Data {
		switch data.Name {
		case "ClientTransactionId":
			subscription.ExternalID = data.Value.(string)
		case "OfferCode":
			subscription.PlanID = data.Value.(string)
		case "SubscriptionStatus":
			subscription.Status = data.Value.(string)
		}
	}
//check subscription status and set the status description accordingly.
	if subscription.Status == "A" {
		subscription.StatusDescription = "Subscriber in active state"
	} else {
		subscription.Status = "D"
		subscription.StatusDescription = "Subscriber in Deactive state"
	}

//query the local database for a subscription matching the external ID and plan ID provided in the notification. 
// NOTE: Better to use partner_id and external_id.
// But plan_id is okay given that it is a 1 to 1 representation of customer.
	if err := database.Db.Debug().Table("subscriptions").Where("external_id = ? plan_id = ?", subscription.ExternalID, subscription.PlanID).First(&subscription).Error; err != nil {
		logrus.Error(err)
		return
	}
// if a matching subscription is found, update the status and status description with the values extracted from the notification
	err := database.Db.Debug().Table("subscriptions").Where("external_id = ? plan_id = ?", subscription.ExternalID, subscription.PlanID).Updates(&subscription).Error
	if err != nil {
		logrus.Error(err)
		return
	}
//marshals the notification payload to JSON and sends a POST request to the subscription's callback URL with the updated information.
	payload, _ := json.Marshal(notification)
	utils.Request(string(payload), map[string][]string{
		"Content-Type": {"application/json"},
	}, subscription.Callback, "POST")
}