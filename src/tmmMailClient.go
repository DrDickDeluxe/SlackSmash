package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"errors"
)

var errTmmStatusCode error = errors.New("TMM responded with invalid HTTP code")

type TmmMail struct {
	TmmSubject	string `json:"subject"`
	TmmBody		string `json:"bodyPlainText"`
}

func (this *TmmMail) Subject() string {
	return this.TmmSubject
}

func (this *TmmMail) Body() string {
	return this.TmmBody
}

const TmmBase = "https://10minutemail.com"

type TmmMailClient struct {
	httpClient		*http.Client
	emailAddress	string
}

func NewTmmMailClient(httpClient *http.Client) *TmmMailClient {
	return &TmmMailClient{
		httpClient: httpClient,
	}
}

const TmmSignupUrl = TmmBase + "/10MinuteMail/resources/session/address"

func (this *TmmMailClient) Signup() error {
	req, _ := http.NewRequest("GET", TmmSignupUrl, nil)
	req.Header.Set("referer", TmmBase + "/")
	req.Header.Set("origin", TmmBase)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errTmmStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	this.emailAddress = string(b)

	return nil
}

func (this *TmmMailClient) Address() string {
	return this.emailAddress
}

const TmmFetchUrl = TmmBase + "/10MinuteMail/resources/messages/messagesAfter/0"

func (this *TmmMailClient) Inbox() ([]Mail, error) {
	req, _ := http.NewRequest("GET", TmmFetchUrl, nil)
	req.Header.Set("referer", TmmBase + "/")
	req.Header.Set("origin", TmmBase)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errTmmStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	var messages []*TmmMail
	if err := json.Unmarshal(b, &messages); err != nil {
		return nil, err
	}
	m := make([]Mail, len(messages))
	for i,v := range messages {
		m[i] = v
	}
	return m, nil
}
