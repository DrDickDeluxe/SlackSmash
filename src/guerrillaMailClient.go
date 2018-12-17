package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"errors"
)

var errGuerrillaStatusCode error = errors.New(
	"Guerrilla responded with an invalid HTTP code")

type GuerMail struct {
	GuerrillaSubject	string `json:"mail_subject"`
	GuerrillaBody		string `json:"mail_body"`
}

func (this *GuerMail) Subject() string {
	return this.GuerrillaSubject
}

func (this *GuerMail) Body() string {
	return this.GuerrillaBody
}

const GuerrillaBase = "http://api.guerrillamail.com"

type GuerrillaMailClient struct {
	httpClient		*http.Client
	emailAddress	string
}

func NewGuerrillaMailClient(httpClient *http.Client) *GuerrillaMailClient {
	return &GuerrillaMailClient{
		httpClient: httpClient,
	}
}

const GuerrillaSignupUrl = GuerrillaBase + "/ajax.php?f=get_email_address"

func (this *GuerrillaMailClient) Signup() error {
	req, _ := http.NewRequest("GET", GuerrillaSignupUrl, nil)
	req.Header.Set("origin", GuerrillaBase)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errGuerrillaStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type Response struct {
		Email	string `json:"email_addr"`
	}
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return err
	}
	this.emailAddress = response.Email

	return nil
}

func (this *GuerrillaMailClient) Address() string {
	return this.emailAddress
}

const GuerrillaGetUrl = GuerrillaBase + "/ajax.php?f=fetch_email"

func (this *GuerrillaMailClient) getEmailById(id int) (*GuerMail, error) {
	req, _ := http.NewRequest("GET",
		fmt.Sprintf("%s&email_id=%d", GuerrillaGetUrl, id), nil)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errGuerrillaStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m GuerMail
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

const GuerrillaFetchUrl = GuerrillaBase + "/ajax.php?f=check_email&seq=0"

func (this *GuerrillaMailClient) Inbox() ([]Mail, error) {
	req, _ := http.NewRequest("GET", GuerrillaFetchUrl, nil)
	req.Header.Set("origin", GuerrillaBase)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errGuerrillaStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type MailData struct {
		Id			int `json:"mail_id"`
		Subject		string `json:"mail_subject"`
	}
	type Response struct {
		List		[]MailData `json:"list"`
	}
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	m := make([]Mail, len(response.List))
	for i,v := range response.List {
		m[i], err = this.getEmailById(v.Id)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}
