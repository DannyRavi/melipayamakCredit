package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os/signal"
	"time"

	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const soapTemplateGetCredit = `<?xml version="1.0" encoding="utf-8"?>
						<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
						<soap12:Body>
							<GetCredit xmlns="http://tempuri.org/">
							<username>VAR1</username>
							<password>VAR2</password>
							</GetCredit>
						</soap12:Body>
						</soap12:Envelope>`

const soapTemplateGetCredit2 = `<?xml version="1.0" encoding="utf-8"?>
								<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
								<soap12:Body>
									<GetUserCredit2 xmlns="http://tempuri.org/">
									<username>VAR1</username>
									<password>VAR2</password>
									</GetUserCredit2>
								</soap12:Body>
								</soap12:Envelope>`

// ----------------------
// Prometheus Metric
// ----------------------
var creditResultIt = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "melliPayamak_get_credit_result_IT",
	Help: "Credit result from SOAP response",
})

// Create a Prometheus gauge metric
var creditResultRial = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "melliPayamak_get_credit_result_Rial",
	Help: "Credit result from SOAP response",
})

// ----------------------
// XML Structs
// ----------------------
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

type Body struct {
	GetCreditResponse      GetCreditResponse      `xml:"GetCreditResponse"`
	GetUserCredit2Response GetUserCredit2Response `xml:"GetUserCredit2Response"`
}

type GetCreditResponse struct {
	GetCreditResult string `xml:"GetCreditResult"`
}

type GetUserCredit2Response struct {
	GetUserCredit2Result string `xml:"GetUserCredit2Result"`
}

// ----------------------
// Interface Definition
// ----------------------
type CreditFetcher interface {
	Fetch() (float64, error)
}

var account smsAcount
var cre creadit
var smsc smsAcount

// ----------------------
// Concrete Implementation
// ----------------------
type SoapCreditFetcher struct {
	source func() (string, error)
}

func (f *SoapCreditFetcher) Fetch() (float64, error) {
	rawXML, err := f.source()
	if err != nil {
		return 0, err
	}

	var envelope Envelope
	decoder := xml.NewDecoder(strings.NewReader(rawXML))
	decoder.DefaultSpace = "http://tempuri.org/" // Namespace-aware parsing

	err = decoder.Decode(&envelope)
	if err != nil {
		return 0, err
	}

	valueStr := envelope.Body.GetCreditResponse.GetCreditResult
	return strconv.ParseFloat(valueStr, 64)
}

// ----------------------
// Metric Updater
// ----------------------
func updateMetric(fetcher CreditFetcher) {
	value, err := fetcher.Fetch()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching credit value:")
		return
	}
	creditResultRial.Set(value)
	log.Debug().Msg("Metric updated with:" + strconv.FormatFloat(value, 'E', -1, 64))
}

type postIt interface {
	Runner() (string, error)
}

type inputData struct {
	url     string
	method  string
	payload string
}

func (i *inputData) Runner() (string, error) {
	client := &http.Client{}
	pl := strings.NewReader(i.payload)
	req, err := http.NewRequest(i.method, i.url, pl)

	if err != nil {
		log.Error().Err(err).Msg("")
		// return
	}
	// "http://tempuri.org/GetCredit"
	// req.Header.Add("SOAPAction", i.header)
	req.Header.Add("Content-Type", "application/soap+xml")

	res, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("")
		// return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("")
		// return
	}
	rec := string(body)
	// fmt.Println()
	return rec, err
}

// Function that uses the interface
func fire(p postIt) string {
	rec, err := p.Runner()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	log.Debug().Msg(rec)
	return rec
}

func decoder(val string, kind string) (float64, error) {
	var envelope Envelope
	decoder := xml.NewDecoder(strings.NewReader(val))
	decoder.DefaultSpace = "http://tempuri.org/" // handles namespace correctly

	err := decoder.Decode(&envelope)
	if err != nil {
		fmt.Println("Error decoding XML:", err)

	}
	valueStr := ""
	switch kind {
	case "GetCredit":
		log.Debug().Str("GetCreditResult:", envelope.Body.GetCreditResponse.GetCreditResult).Send()
		valueStr = envelope.Body.GetCreditResponse.GetCreditResult

	case "GetUserCredit2":
		log.Debug().Str("GetCreditResult2:", envelope.Body.GetUserCredit2Response.GetUserCredit2Result).Send()
		valueStr = envelope.Body.GetUserCredit2Response.GetUserCredit2Result
	default:
		log.Error().Msg("mistake kind value")
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Error().Err(err).Msg("Error converting value: ")
	}
	// creditResult.Set(value)
	log.Info().Msg("Updated metric with value:" + strconv.FormatFloat(value, 'E', -1, 64))
	return value, err
}

type smsAcount struct {
	User          []string
	Pass          []string
	payload       []string
	interval_time int
}

// var smsAcc smsAcount

// type theSMS struct {
// 	smsAcount []smsAcount
// }

type creadit struct {
	totoal_melipayamk float64
	user_meliPayamak  float64
	// payload []string
}

func fillSms(_url string, _method string, _payload string) inputData {
	__url__ := "http://api.payamak-panel.com/" + _url
	return inputData{
		url:     __url__,
		method:  _method,
		payload: _payload,
	}
}

func recSMSCost(inp inputData, theKind string) float64 {

	rx0 := fire(&inp)
	rx, err := decoder(rx0, theKind)
	if err != nil {
		fmt.Println(err)
	}
	return rx

}

func templateFill(account smsAcount, _user []string, _pass []string, template string) (acc smsAcount) {
	account = smsAcount{}
	count := len(_user)

	for i := range count {

		account.User = append(account.User, _user[i])
		account.Pass = append(account.Pass, _pass[i])
		// xmlString := strings.Replace(soapTemplateGetCredit, "VAR1", users[i], 1)
		// strings.Replace(soapTemplateGetCredit, "VAR2", pass[i], 1)
		// Create a replacer that replaces VAR1, VAR2, VAR3 in a single pass
		replacer := strings.NewReplacer(
			"VAR1", _user[i],
			"VAR2", _pass[i],
		)
		soapString := replacer.Replace(template)
		// fmt.Println(soapString)
		account.payload = append(account.payload, soapString)
	}
	return account

}

func setLogLevel(getTheEnv string) {
	switch getTheEnv {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

}

func updateMetrics(accSms smsAcount, underTest bool) float64 {
	if underTest {
		for {

			theAcc := templateFill(account, accSms.User, accSms.Pass, soapTemplateGetCredit)
			rec := fillSms("post/Send.asmx", "POST", theAcc.payload[0])
			cre.user_meliPayamak = recSMSCost(rec, "GetCredit")
			log.Warn().Msg(strconv.FormatFloat(cre.user_meliPayamak, 'E', -1, 64))

			theAcc = templateFill(account, accSms.User, accSms.Pass, soapTemplateGetCredit2)
			rec = fillSms("post/Users.asmx", "POST", theAcc.payload[1])
			cre.totoal_melipayamk = recSMSCost(rec, "GetUserCredit2")
			log.Warn().Msg(strconv.FormatFloat(cre.totoal_melipayamk, 'E', -1, 64))

			creditResultRial.Set(cre.totoal_melipayamk)
			creditResultIt.Set(cre.user_meliPayamak)
			time.Sleep(time.Duration(accSms.interval_time) * time.Minute)
		}
	}
	theAcc := templateFill(account, accSms.User, accSms.Pass, soapTemplateGetCredit2)
	rec := fillSms("post/Users.asmx", "POST", theAcc.payload[1])
	cre.totoal_melipayamk = recSMSCost(rec, "GetUserCredit2")
	log.Warn().Msg(strconv.FormatFloat(cre.totoal_melipayamk, 'E', -1, 64))

	return cre.totoal_melipayamk
}

// ----------------------
// Main
// ----------------------

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	prometheus.MustRegister(creditResultRial)
	prometheus.MustRegister(creditResultIt)
	err := godotenv.Load()

	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}

	// Parse comma-separated string into slice
	users := strings.Split(os.Getenv("users"), ",")
	pass := strings.Split(os.Getenv("pass"), ",")
	log_level := os.Getenv("log_level")
	interval_time, err := strconv.Atoi(os.Getenv("interval_time"))
	if err != nil {
		log.Fatal().Msg("Error:" + err.Error())
	}
	setLogLevel(log_level)
	if len(users) != len(pass) {
		log.Fatal().Msg("users and pass have different lengths")
	}
	for i := range users {
		log.Debug().Str("users Origins:", users[i]).Send()
		log.Debug().Str("pass:", pass[i]).Send()

		smsc.User = append(smsc.User, users[i])
		smsc.Pass = append(smsc.Pass, pass[i])

	}
	smsc.interval_time = interval_time

}
func main() {

	// Loop through arrays and populate struct

	log.Info().Msg("----Servers loaded----")
	go updateMetrics(smsc, true)

	http.Handle("/metrics", promhttp.Handler())
	log.Warn().Msgf("Serving metrics on :8285/metrics")

	// Create an http.Server manually
	server := &http.Server{
		Addr:    ":8285",
		Handler: nil, // nil uses http.DefaultServeMux
	}

	// varz := http.ListenAndServe(":8285", nil)

	// log.Fatal().Msg(varz.Error())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt) // listen for SIGINT (Ctrl+C)

	// // Run server in a goroutine so it doesn’t block
	go func() {
		log.Warn().Msgf("Starting server on :8285")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Warn().Msgf("Could not listen: %v\n", err)
		}
	}()

	// Block until a signal is received
	<-stop
	log.Warn().Msg("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown gracefully
	if err := server.Shutdown(ctx); err != nil {
		// msg := fmt.Sprintf("Server forced to shutdown: %v", err)
		log.Warn().Msgf("Server forced to shutdown: %v", err)
	}

	log.Warn().Msg("Server exited gracefully")
}
