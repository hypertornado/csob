package csob

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CSOB struct {
	merchantId         string
	key                *rsa.PrivateKey
	testingEnvironment bool
	client             *http.Client
	returnUrl          string
}

func NewCSOBTestingEnvironment(merchantId, privateKeyPath string) (*CSOB, error) {
	csob, err := NewCSOB(merchantId, privateKeyPath, "http://www.example.com")
	if err == nil {
		csob.testingEnvironment = true
	}
	return csob, err
}

func NewCSOB(merchantId, privateKeyPath, returnUrl string) (*CSOB, error) {
	key, err := loadKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &CSOB{
		merchantId:         merchantId,
		key:                key,
		testingEnvironment: false,
		client:             &http.Client{},
		returnUrl:          returnUrl,
	}, nil
}

func (c *CSOB) Echo() error {
	dateStr := timestamp()

	signature, err := c.sign(c.merchantId, dateStr)
	if err != nil {
		return err
	}

	params := map[string]interface{}{
		"dttm":       dateStr,
		"signature":  signature,
		"merchantId": c.merchantId,
	}

	resp, err := c.apiRequest("POST", "/echo", params)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	return nil
}

type PaymentResult struct {
	PayId         string `json:"payId"`
	Dttm          string `json:"dttm"`
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	PaymentStatus int    `json:"paymentStatus"`
	AuthCode      string `json:"authCode"`
	Signature     string `json:"signature"`
}

/*func (c *CSOB) IsResultValid(paymentResult *PaymentResult) bool {
	signature, err := c.sign(
		paymentResult.PayId,
		paymentResult.Dttm,
		fmt.Sprintf("%d", paymentResult.ResultCode),
		paymentResult.ResultMessage,
	)
	println(signature)
	println(paymentResult.Signature)
	if err != nil {
		return false
	}
	if signature == paymentResult.Signature {
		return true
	}
	return false
}*/

func (c *CSOB) Init(orderNo, name string, quantity, amount uint, description string) (*PaymentResult, error) {

	amountStr := fmt.Sprintf("%d", amount)
	quantityStr := fmt.Sprintf("%d", quantity)

	total := amountStr

	closePayment := "true"

	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      orderNo,
		"dttm":         timestamp(),
		"payOperation": "payment",
		"payMethod":    "card",
		"totalAmount":  total,
		"currency":     "CZK",
		"closePayment": closePayment,
		"returnUrl":    c.returnUrl,
		"returnMethod": "POST",
		"cart": []interface{}{
			map[string]interface{}{
				"name":        name,
				"quantity":    quantityStr,
				"amount":      amountStr,
				"description": description,
			},
		},
		"description":  description,
		"merchantData": nil,
		"customerId":   nil,
		"language":     "CZ",
		"signature":    "",
	}

	signature, err := c.sign(
		c.merchantId,
		params["orderNo"],
		params["dttm"],
		"payment",
		"card",
		total,
		"CZK",
		closePayment,
		c.returnUrl,
		"POST",
		name,
		quantityStr,
		amountStr,
		description,
		description,
		"CZ",
	)
	if err != nil {
		return nil, err
	}

	params["signature"] = signature

	resp, err := c.apiRequest("POST", "/payment/init", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, csobError
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentResult PaymentResult
	err = json.Unmarshal(respBytes, &paymentResult)
	return &paymentResult, err
}
