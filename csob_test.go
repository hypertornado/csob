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

/*
func TestResponseSignature(t *testing.T) {
	crypted := "pJP6UCBqB7Wduc2L3zc2xC+Uqb1fQlwXTyLXsdcbo3mM+UtQW2yPKYMuXu62YGAvuLM7x+l9wCeHRZts2xhki+rkAtQDw59SNNf0dhpIqVkSDqjcscV0lv+7y+HEmXzibM+4VvclITmB4pIHdKcPkaX3iRzpBdk46INlzbZoG+KoP4s+Xp/tqYrKA0pUn+y9s6P08U+tlo3vJl2BTtflldDHzBlDXJaHGvfWG/G+I1i7ToiDnSG0BILeODaNlCL6YyJjnQzHiNzd0F4IVSjpaufCdHAxLDrF4mBybGehhDF1i/3xFLMGRj4Ct2B+mmJ6jdKLJrUQ/CMbZEHTOyKjAQ=="
	data := "20161013094628|0|OK"

	key, err := loadKey(testKeyPath() + "/rsa.key")
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := decrypt(key, crypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != data {
		t.Fatal(decrypted)
	}
}
*/

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
	csob, _ := prepareTest()

	order := csob.NewOrder(1234, "some name", "some description")
	order.AddItem("item name", 1, 200000)
	order.Close()

	resp, err := csob.Init(order)
	if err != nil {
		t.Error(err)
	}

	if resp.ResultMessage != "OK" {
		t.Error(resp.ResultMessage)
	}

	if resp.PaymentStatus != 1 {
		t.Error("error")
	}

	_, err = csob.ProcessURL(resp)
	if err != nil {
		t.Error(err)
	}

	status, err := csob.Status(resp.PayId)
	if err != nil {
		t.Fatal(err)
	}
	if status.ResultMessage != "OK" {
		t.Error("error")
	}

}

func testKeyPath() string {
	return os.Getenv("HOME") + "/.csob/test_keys"
}

func prepareTest() (*CSOB, error) {

	keyPath := testKeyPath()

	data, err := ioutil.ReadFile(keyPath + "/merchantId.txt")
	if err != nil {
		return nil, err
	}

	c, err := NewCSOB(string(data), keyPath+"/rsa.key")
	if err != nil {
		return nil, err
	}

	c.ReturnUrl("GET", "http://www.example.com")
	c.TestingEnvironment()
	return c, nil

}
