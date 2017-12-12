package deepsecurity

import (
	"crypto/tls"
	"github.com/jeffthorne/deepsecurity-go/gowsdlservice"
	"fmt"
	"log"
	"net/http"
	"github.com/levigross/grequests"
)

type DSM struct {
	SessionID  string
	Host       string
	Port       string
	Tenant     string
	RestURL    string
	RestClient http.Client
	SoapClient *gowsdlservice.Manager
	SoapURL    string
}

var dsasHost string = "app.deepsecurity.trendmicro.com"
var dsasPort string = "443"


func NewDSM(username string, password string, host string, port string, tenant string, verifySSL bool) DSM {
	dsm := DSM{Host: host, Port: port, Tenant: tenant}
	if dsm.Host == "" {
		dsm.Host = dsasHost
		dsm.Port = dsasPort
	}

	dsm.RestURL = fmt.Sprintf("https://%s:%s/rest/", dsm.Host, dsm.Port)
	dsm.SoapURL = fmt.Sprintf("https://%s:%s/webservice/Manager", dsm.Host, dsm.Port)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !verifySSL}}
	dsm.RestClient = http.Client{Transport: tr}
	authenticate(username, password, &dsm, verifySSL)
	return dsm
}

func authenticate(username string, password string, dsm *DSM, verifySSL bool) {

	var str string
	auth := gowsdlservice.BasicAuth{Login: username, Password: password}
	dsm.SoapClient = gowsdlservice.NewManager(dsm.SoapURL, true, &auth)

	if dsm.Tenant != "" {
		tenantAuth := gowsdlservice.AuthenticateTenant{TenantName: dsm.Tenant, Username: username, Password: password}
		str = fmt.Sprintf("{\"dsCredentials\":{\"tenantName\": \"%s\", \"password\": \"%s\", \"userName\": \"%s\"}}", dsm.Tenant, password, username)
		authResponse, err := dsm.SoapClient.AuthenticateTenant(&tenantAuth)
		if err != nil {
			log.Fatal(err)
		}
		dsm.SessionID = authResponse.AuthenticateTenantReturn
	} else {
		str = fmt.Sprintf("{\"dsCredentials\":{\"password\": \"%s\", \"userName\": \"%s\"}}", password, username)
		url := fmt.Sprintf("%sauthentication/login", dsm.RestURL)
		res, err := grequests.Post(url, &grequests.RequestOptions{HTTPClient: &dsm.RestClient, JSON: []byte(str), IsAjax: false})
		if err != nil {
			log.Fatalf("There was a problem %v", err)
		}
		dsm.SessionID = res.String()
	}

}

func (dsm DSM) EndSession() {
	url := fmt.Sprintf("%sauthentication/logout", dsm.RestURL)
	_, err := grequests.Delete(url, &grequests.RequestOptions{HTTPClient: &dsm.RestClient, Params: map[string]string{"sID": dsm.SessionID}})

	if err != nil {
		log.Println("Unable to make request", err)
	}
}


func (dsm DSM)HostClearWarningsErrors(hosts []int32) *gowsdlservice.HostClearWarningsErrorsResponse{
	hce := gowsdlservice.HostClearWarningsErrors{HostIDs:hosts, SID:dsm.SessionID,}
	response, _ := dsm.SoapClient.HostClearWarningsErrors(&hce)
	return response
}

func (dsm DSM) HostGetStatus(host int32) *gowsdlservice.HostGetStatusResponse {
	hgs := gowsdlservice.HostGetStatus{Id: host, SID: dsm.SessionID,}
	response, err := dsm.SoapClient.HostGetStatus(&hgs)

	if err != nil {
		fmt.Println("in nill", err)

	}

	return response
}



func (dsm DSM) HostRecommendationScan(hosts []int32){
	hrs := gowsdlservice.HostRecommendationScan{HostIDs: hosts, SID: dsm.SessionID}
	response, err := dsm.SoapClient.HostRecommendationScan(&hrs)
	if err != nil{
		log.Println("Error Initiating reccomentation scan:", err)
	}
	fmt.Println(response)

}


func (dsm DSM)HostRetrieveByName(hostName string) *gowsdlservice.HostTransport{

	hrbn := gowsdlservice.HostRetrieveByName{Hostname: hostName, SID:dsm.SessionID}
	resp, err := dsm.SoapClient.HostRetrieveByName(&hrbn)
	if err != nil{
		log.Println("Error retrieving host:", err)
	}

	hostTransPort := resp.HostRetrieveByNameReturn
	return hostTransPort
}






