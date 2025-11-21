package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// func TestAdd(t *testing.T) {
//     result := Add(2, 3)
//     if result != 5 {
//         t.Errorf("Add(2, 3) = %d; expected 5", result)
//     }
// }

func TestLoadEnv(t *testing.T) {

	err := godotenv.Load()

	if err != nil {
		t.Errorf("Error loading .env file")
	}

	test := strings.Split(os.Getenv("test"), ",")

	if test[0] != "ok" {
		t.Errorf("mistake to read Env File")
	}
}

func TestMelipayaMAkStatus(t *testing.T) {

	url := "https://rest.payamak-panel.com/api/SendSMS/GetCredit"
	method := "POST"

	payload := strings.NewReader(``)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		t.Errorf("melipayamak or network is down")
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		t.Errorf("http client side error")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Errorf("tatus code not valid")

	}

}

func TestFunctionality(t *testing.T) {
	flo := updateMetrics(smsc,false)
	pi := 3.1415926535
	ans := flo - pi
	fmt.Println(ans)
	if ans < 0 {
		t.Errorf("miss calculate value")
	}

}
