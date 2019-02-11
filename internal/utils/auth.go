package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type oidcResponse struct {
	AccessToken string `json:"access_token"`
}

// GetServiceAccountToken retrieves service accout token from the OIDC provider
func GetServiceAccountToken(serviceAccountID, serviceAccountSecret, oidcProvider string) string {
	formURL := fmt.Sprintf("%s/token", oidcProvider)
	resp, err := http.PostForm(formURL,
		url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {serviceAccountID},
			"client_secret": {serviceAccountSecret},
		})
	if err != nil {
		log.Fatalf("Could not get service account token")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body when getting service account token")
	}

	var token oidcResponse
	json.Unmarshal(body, &token)
	return token.AccessToken
}
