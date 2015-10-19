package csob

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEcho(t *testing.T) {
	csob, _ := prepareTest()
	err := csob.Echo()
	if err != nil {
		t.Error(err)
	}
}

func TestEchoGet(t *testing.T) {
	csob, _ := prepareTest()
	err := csob.EchoGet()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSignature(t *testing.T) {
	data := "A1233aBcVn|554822|20151013133307|payment|card|1789600|CZK|true|https://vasobchod.cz/return-gateway|POST|žluťoučký kůň|1|1789600|Lenovo ThinkPad Edge E540|Poštovné|1|0|Doprava PPL|Nákup žluťoučký kůň na vasobchod.cz (Lenovo ThinkPad Edge E540, Doprava PPL)|c29tZS1iYXNlNjQtZW5jb2RlZC1tZXJjaGFudC1kYXRh|CZ"
	expectedSignature := "tGgxtGCiGqxi6isgNJUk8A02pJQQ/E7aOcafJz/alKYZajD3yiB5bGDS6njVzoNOcwgNlVrhwPhXlzKPGPzg56NhSIE/EvBEqkJF/Y950e8YpJGHzoXuf90HqMlJ0Sq5c8W/jRnWGshf8uVzxd7obMZOdcHXmVOOxkAQyyoIUhsmOEVnfIjy26YT87evIsGSSH263LdScK2JpbDdhOQk2Lfcypil0bXFdnzGaSHaTRbtPovcLxrkFA1r3ey6ntGfphi72kDij+Xr+zZPHuuU3VAQZ/xAIWsFpW8XmQam5YIPOrAqHNgNcv+ojvWtYl35l6FeDlmD/HIzc2AdD6Offg=="

	key, err := loadKey(testKeyPath() + "/rsa.key")
	if err != nil {
		t.Error(err)
	}
	signature, err := signData(key, data)
	if err != nil {
		t.Error(err)
	}

	if expectedSignature != signature {
		t.Error(signature)
	}
}

func TestInitPayment(t *testing.T) {

	paymentId := timestamp()[4:]

	csob, _ := prepareTest()
	resp, err := csob.Init(paymentId, "some name", 2546, 200, "some description")
	if err != nil {
		t.Error(err)
	}

	if resp.ResultMessage != "OK" {
		t.Error("error")
	}

	if resp.PaymentStatus != 1 {
		t.Error("error")
	}

	_, err = csob.ProcessURL(resp)
	if err != nil {
		t.Error(err)
	}

	url, err := csob.ProcessURL(resp)
	if err != nil {
		t.Error(err)
	}
	println(url)

}

func testKeyPath() string {
	return os.Getenv("HOME") + "/.csob/test_keys"
}

func prepareTest() (*CSOB, error) {

	keyPath := testKeyPath()

	data, err := ioutil.ReadFile(keyPath + "/merchantId.txt")
	if err != nil {
		panic(err)
	}

	return NewCSOBTestingEnvironment(string(data), keyPath+"/rsa.key")
}
