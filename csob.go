package csob

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func HumanReadableStatusCzech(id int) string {
	switch id {
	case 1:
		return "Platba založena"
	case 2:
		return "Platba probíhá"
	case 3:
		return "Platba zrušena"
	case 4:
		return "Platba potvrzena"
	case 5:
		return "Platba odvolána"
	case 6:
		return "Platba zamítnuta"
	case 7:
		return "Čekání na zůčtování"
	case 8:
		return "Platba zůčtována"
	case 9:
		return "Zpracování vrácení"
	case 10:
		return "Platba vrácena"
	}
	return "Žádný stav"
}

func NewCSOB(merchantId, privateKeyPath string) (*CSOB, error) {
	key, err := loadKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return &CSOB{
		merchantId:         merchantId,
		key:                key,
		testingEnvironment: false,
		client:             &http.Client{},
	}, nil
}

func (c *CSOB) TestingEnvironment() {
	c.testingEnvironment = true
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
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return csobError
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var echoResp echoResponse
	err = json.Unmarshal(respBytes, &echoResp)
	if err != nil {
		return err
	}

	//TODO: check signature

	return nil
}

type echoResponse struct {
	Dttm          string `json:"dttm"`
	ResultCode    int    `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	Signature     string `json:"signature"`
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
	defer resp.Body.Close()

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

func (c *CSOB) Close(payId string) error {
	return c.paymentStatusTypePutCall(payId, "close")
}

func (c *CSOB) Reverse(payId string) error {
	return c.paymentStatusTypePutCall(payId, "reverse")
}

func (c *CSOB) Refund(payId string) error {
	return c.paymentStatusTypePutCall(payId, "refund")
}

type CSOB struct {
	merchantId         string
	key                *rsa.PrivateKey
	testingEnvironment bool
	client             *http.Client
	returnUrl          string
	returnMethod       string
}

func (c *CSOB) ReturnUrl(returnMethod, returnUrl string) {
	c.returnMethod = returnMethod
	c.returnUrl = returnUrl
}

type orderItem struct {
	name     string
	quantity uint
	amount   uint
}

type order struct {
	orderNo      uint
	name         string
	description  string
	quantity     uint
	amount       uint
	closePayment bool
	orderItems   []orderItem
	payOperation string
	payMethod    string
	language     string
	currency     string
}

func (c *CSOB) NewOrder(orderNo uint, name, description string) *order {
	return &order{
		orderNo:      orderNo,
		name:         name,
		description:  description,
		closePayment: false,
		orderItems:   []orderItem{},
		payOperation: "payment",
		payMethod:    "card",
		language:     "EN",
		currency:     "USD",
	}
}

func (o *order) Close() {
	o.closePayment = true
}

func (o *order) Language(language string) {
	o.language = language
}

func (o *order) Currency(currency string) {
	o.currency = currency
}

func (o *order) AddItem(name string, quantity, amount uint) {
	var item = orderItem{
		name:     name,
		quantity: quantity,
		amount:   amount,
	}
	o.orderItems = append(o.orderItems, item)

}

func (c *CSOB) Init(order *order) (*PaymentStatus, error) {

	if len(order.orderItems) != 1 {
		return nil, errors.New("there should be 1 order item")
	}

	amountStr := fmt.Sprintf("%d", order.orderItems[0].amount)
	quantityStr := fmt.Sprintf("%d", order.orderItems[0].quantity)

	total := amountStr

	closePaymentStr := "true"
	if order.closePayment {
		closePaymentStr = "false"
	}

	orderNoStr := fmt.Sprintf("%d", order.orderNo)

	params := map[string]interface{}{
		"merchantId":   c.merchantId,
		"orderNo":      orderNoStr,
		"dttm":         timestamp(),
		"payOperation": order.payOperation,
		"payMethod":    order.payMethod,
		"totalAmount":  total,
		"currency":     order.currency,
		"closePayment": closePaymentStr,
		"returnUrl":    c.returnUrl,
		"returnMethod": c.returnMethod,
		"cart": []interface{}{
			map[string]interface{}{
				"name":     order.orderItems[0].name,
				"quantity": quantityStr,
				"amount":   amountStr,
			},
		},
		"description":  order.description,
		"merchantData": nil,
		"customerId":   nil,
		"language":     order.language,
		"signature":    "",
	}

	signature, err := c.sign(
		c.merchantId,
		orderNoStr,
		params["dttm"],
		order.payOperation,
		order.payMethod,
		total,
		order.currency,
		closePaymentStr,
		c.returnUrl,
		c.returnMethod,
		order.orderItems[0].name,
		quantityStr,
		amountStr,
		order.description,
		order.language,
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
