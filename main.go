package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type requestPayload struct {
	ProxyCondition string `json:"proxy_condition"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getListenAddress() string {
	port := getEnv("PORT", "1338")
	return ":" + port
}

func logSetup() {
	a_condition_url := os.Getenv("A_CONDITION_URL")
	b_condition_url := os.Getenv("B_CONDITION_URL")
	default_condition_url := os.Getenv("DEFAULT_CONDITION_URL")

	log.Println("Server will run on: ", getListenAddress())
	log.Println("Redirecting to A url: ", a_condition_url)
	log.Println("Redirecting to B url: ", b_condition_url)
	log.Println("Redirecting to default url: ", default_condition_url)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request){
	requestPayload := parseRequestBody(req)
	url := getProxyUrl(requestPayload.ProxyCondition)

	serveReverseProxy(url, res, req)
}

func parseRequestBody(req *http.Request) requestPayload{
	var requestPayload requestPayload

	err := json.NewDecoder(req.Body).Decode(&requestPayload)
	if err != nil {
		panic(err)
	}

	return requestPayload
}

func getProxyUrl(proxyConditionRaw string) string {
	proxyCondition := strings.ToUpper(proxyConditionRaw)

	a_condition_url := os.Getenv("A_CONDITION_URL")
	b_condition_url := os.Getenv("B_CONDITION_URL")
	default_condition_url := os.Getenv("DEFAULT_CONDITION_URL")

	if proxyCondition == "A" {
		return a_condition_url
	}

	if proxyCondition == "B" {
		return b_condition_url
	}

	return default_condition_url
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request){
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forward-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

func main(){
	logSetup()

	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}