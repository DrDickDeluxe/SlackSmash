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

var errLaraInvalidStatusCode error = errors.New("LaraChat invalid status code")
var errLaraNoCsrf error = errors.New("Failed to obtain LaraChat CSRF token")

type LaraChatAccountCreator struct {
	InviteUrl		string
	mailClient		MailClient
	httpClient		*http.Client
	baseUrl			string
	inviteCode		string
	csrfToken		string
}

func NewLaraChatAccountCreator(
		inviteUrl string,
		mailClient MailClient,
		httpClient *http.Client) *LaraChatAccountCreator {
	return &LaraChatAccountCreator{
		InviteUrl: inviteUrl,
		mailClient: mailClient,
		httpClient: httpClient,
	}
}

func (this *LaraChatAccountCreator) ApiBaseUrl() string {
	return this.baseUrl
}

func (this *LaraChatAccountCreator) InviteCode() string {
	return this.inviteCode
}

var reLaraInput *regexp.Regexp = regexp.MustCompile(
	`<input[^>]+name=\"([^\">]+)\"[^>]+value=\"([^>]*)\"\s*>`)

func (this *LaraChatAccountCreator) getCsrfToken() error {
	req, _ := http.NewRequest("GET", this.InviteUrl, nil)
	req.Header.Set("referer", "https://larachat.co/")
	req.Header.Set("origin", "https://larachat.co")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errLaraInvalidStatusCode
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	dom := string(b)

	capture := reLaraInput.FindStringSubmatch(dom)
	if len(capture) < 3 {
		return errLaraNoCsrf
	}
	this.csrfToken = capture[2]

	return nil
}

func (this *LaraChatAccountCreator) submitForm() error {
	form := url.Values{}
	form.Add("_token", this.csrfToken)
	form.Add("first_name", "Hello")
	form.Add("last_name", "World")
	form.Add("email", this.mailClient.Address())
	form.Add("newsletter", "1")
	req, _ := http.NewRequest("POST",
		this.InviteUrl, strings.NewReader(form.Encode()))
	req.Header.Set("referer", this.InviteUrl)
	req.Header.Set("origin", "https://larachat.co")
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errLaraInvalidStatusCode
	}
	return nil
}

func (this *LaraChatAccountCreator) Create(name string) error {
	if err := this.getCsrfToken(); err != nil {
		return err
	}
	if err := this.submitForm(); err != nil {
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
