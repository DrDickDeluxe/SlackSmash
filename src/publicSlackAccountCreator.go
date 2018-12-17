package main

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"regexp"
	"time"
	"strings"
	"errors"
)

type PublicSlackAccountCreator struct {
	InviteUrl		string
	mailClient		MailClient
	httpClient		*http.Client
	baseUrl			string
	inviteCode		string
}

var errPsStatusCode error = errors.New(
	"PublicSlack responded with invalid HTTP code")

func NewPublicSlackAccountCreator(
		inviteUrl string,
		mailClient MailClient,
		httpClient *http.Client) *PublicSlackAccountCreator {
	return &PublicSlackAccountCreator{
		InviteUrl: inviteUrl,
		mailClient: mailClient,
		httpClient: httpClient,
	}
}

func (this *PublicSlackAccountCreator) ApiBaseUrl() string {
	return this.baseUrl
}

func (this *PublicSlackAccountCreator) InviteCode() string {
	return this.inviteCode
}

const RegexInputs = `<input[^>]+name=\"([^\">]+)\"[^>]+value=\"([^>]*)\"\s+\/>`
var rePsInput = regexp.MustCompile(RegexInputs)

func (this *PublicSlackAccountCreator) getSlackInviteEmail() error {
	req, _ := http.NewRequest("GET", this.InviteUrl, nil)
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return errPsStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	dom := string(b)

	formInputs := rePsInput.FindAllStringSubmatch(dom, -1)

	form := url.Values{}
	for _,v := range formInputs {
		form.Add(v[1], v[2])
	}
	form.Add("invite[slack_id]", "3")
	form.Add("invite[email]", this.mailClient.Address())
	req, _ = http.NewRequest("POST",
		strings.TrimRight(this.InviteUrl, "/new"),
		strings.NewReader(form.Encode()))
	resp, err = this.httpClient.Do(req)
	if err != nil {
		return err
	}

	resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errPsStatusCode
	}

	return nil
}

func (this *PublicSlackAccountCreator) Create(name string) error {
	if err := this.getSlackInviteEmail(); err != nil {
		return err
	}
	var err error
	this.baseUrl, this.inviteCode, err = PollInboxForSlackMessage(
		this.mailClient, time.Minute * 5)
	if err != nil {
		return err
	}

	return nil
}
