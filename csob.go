package csob

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
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

func (c *CSOB) EchoGet() error {
	dttm := timestamp()
	signature, err := c.sign(c.merchantId, dttm)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/echo/%s/%s/%s",
		c.merchantId,
		dttm,
		url.QueryEscape(signature),
	)
	resp, err := c.apiRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	return nil
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

type PaymentStatus struct {
	PayId         string `json:"payId"`
	Dttm          string `json:"dttm"`
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	PaymentStatus int    `json:"paymentStatus"`
	AuthCode      string `json:"authCode"`
	Signature     string `json:"signature"`
}

func (c *CSOB) ProcessURL(paymentStatus *PaymentStatus) (string, error) {
	dttm := timestamp()
	signature, err := c.sign(c.merchantId, paymentStatus.PayId, dttm)
	if err != nil {
		return "", err
	}

	ret := fmt.Sprintf("%s/payment/process/%s/%s/%s/%s",
		c.baseUrl(),
		c.merchantId,
		paymentStatus.PayId,
		dttm,
		url.QueryEscape(signature),
	)
	return ret, nil
}

func (c *CSOB) Status(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "GET", "payment/status")
}

func (c *CSOB) Reverse(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/reverse")
}

func (c *CSOB) Close(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/close")
}

func (c *CSOB) Refund(payId string) (*PaymentStatus, error) {
	return c.paymentStatusTypeCall(payId, "PUT", "payment/refund")
}

/*type Order struct {
	OrderNo string
}*/

func (c *CSOB) Init(orderNo, name string, quantity, amount uint, description string, closePayment bool) (*PaymentStatus, error) {

	amountStr := fmt.Sprintf("%d", amount)
	quantityStr := fmt.Sprintf("%d", quantity)

	total := amountStr

	closePaymentStr := "true"
	if closePayment {
		closePaymentStr = "false"
	}

	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      orderNo,
		"dttm":         timestamp(),
		"payOperation": "payment",
		"payMethod":    "card",
		"totalAmount":  total,
		"currency":     "CZK",
		"closePayment": closePaymentStr,
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
		closePaymentStr,
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

	return parseStatusResponse(resp)

}
