package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
	"golang.org/x/net/proxy"
)

type Account struct {
	DisplayName		string
	mailClient		MailClient
	accountCreator	AccountCreator
	httpClient		*http.Client
	slackTeamId		string
	apiToken		string
	userId			string
	channels		map[string]string
}

var errInvalidStatusCode error = errors.New("Invalid status code")
var errNoTeamId error = errors.New("Slack invite URL does not have Team ID")
var errSlackApi error = errors.New("Slack API request failed")
var errInvalidChannel error = errors.New("Slack channel does not exist in map.")

func NewAccount(publicInvite string, prx string, displayName string,
		tm MailClientType, ta AccountCreatorType) *Account {
	cookieJar, _ := cookiejar.New(nil)
	dialer, _ := proxy.SOCKS5("tcp", prx, nil, proxy.Direct)
	httpClient := &http.Client{
		Jar: cookieJar,
		Transport: &http.Transport{
			Dial: dialer.Dial,
			IdleConnTimeout: time.Second * 60,
		},
	}

	mailClient := NewMailClient(tm, httpClient)
	accountCreator := NewAccountCreator(ta, publicInvite,
		mailClient, httpClient)

	return &Account{
		DisplayName: displayName,
		mailClient: mailClient,
		accountCreator: accountCreator,
		httpClient: httpClient,
		channels: make(map[string]string),
	}
}

var reTeam *regexp.Regexp = regexp.MustCompile(
	`TS\.web\.invite\.encoded_team_id = \"([\w\d]+)\";`)

func (this *Account) getSlackTeamId() error {
	u := fmt.Sprintf("%s/join/invite/%s",
		this.accountCreator.ApiBaseUrl(),
		this.accountCreator.InviteCode())
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Origin", this.accountCreator.ApiBaseUrl())
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 302 {
		return errInvalidStatusCode
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	dom := string(b)

	capture := reTeam.FindStringSubmatch(dom)
	if len(capture) == 0 {
		return errNoTeamId
	}
	this.slackTeamId = capture[1]

	return nil
}

const ApiCreateUser = "/api/signup.createUser"

var reCode *regexp.Regexp = regexp.MustCompile(`invite\/(.*)\?`)

func (this *Account) finishCreatingSlackAccount() error {
	form := url.Values{}
	form.Add("code", this.accountCreator.InviteCode())
	form.Add("invite_type", "")
	form.Add("team", this.slackTeamId)
	form.Add("tz", "America/New_York")
	form.Add("password", "nopenope12")
	form.Add("emailok", "true")
	form.Add("real_name", this.DisplayName)
	form.Add("display_name", this.DisplayName)
	form.Add("locale", "en-US")
	form.Add("last_tos_acknowledged", "tos_mar2018")
	req, _ := http.NewRequest("POST",
		this.accountCreator.ApiBaseUrl() + ApiCreateUser,
		strings.NewReader(form.Encode()))
	req.Header.Set("Origin", this.accountCreator.ApiBaseUrl())
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errInvalidStatusCode
	}

	type Response struct {
		Ok       bool   `json:"ok"`
		UserId   string `json:"user_id"`
		ApiToken string `json:"api_token"`
	}
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return err
	}

	if !response.Ok {
		return errSlackApi
	}

	this.userId = response.UserId
	this.apiToken = response.ApiToken

	return nil
}

const ApiListChannels = "/api/conversations.list"

func (this *Account) loadChannelList() error {
	form := url.Values{}

	form.Add("exclude_members", "1")
	form.Add("types", "public_channel")
	form.Add("limit", "1000")
	form.Add("include_shared", "true")
	form.Add("token", this.apiToken)
	form.Add("_x_mode", "online")
	req, _ := http.NewRequest("POST",
		this.accountCreator.ApiBaseUrl() + ApiListChannels,
		strings.NewReader(form.Encode()))
	req.Header.Set("Origin", this.accountCreator.ApiBaseUrl())
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	b, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errInvalidStatusCode
	}

	type Channel struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	type Response struct {
		Ok       bool      `json:"ok"`
		Channels []Channel `json:"channels"`
	}
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return err
	}
	for _, v := range response.Channels {
		this.channels[v.Name] = v.Id
	}

	return nil
}

func (this *Account) Signup() error {
	if err := this.mailClient.Signup(); err != nil {
		return err
	}
	fmt.Println("Got email", this.mailClient.Address())
	if err := this.accountCreator.Create(this.DisplayName); err != nil {
		return err
	}
	if err := this.getSlackTeamId(); err != nil {
		return err
	}
	if err := this.finishCreatingSlackAccount(); err != nil {
		return err
	}
	fmt.Println("Joined slack")
	if err := this.loadChannelList(); err != nil {
		return err
	}
	return nil
}

const ApiJoinChannel = "/api/conversations.join"

func (this *Account) JoinChannel(channel string) error {
	form := url.Values{}

	if _, ok := this.channels[channel]; !ok {
		return errInvalidChannel
	}

	form.Add("channel", this.channels[channel])
	form.Add("_in_background", "false")
	form.Add("token", this.apiToken)
	form.Add("_x_mode", "online")
	req, _ := http.NewRequest("POST",
		this.accountCreator.ApiBaseUrl() + ApiJoinChannel,
		strings.NewReader(form.Encode()))
	req.Header.Set("Origin", this.accountCreator.ApiBaseUrl())
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errInvalidStatusCode
	}

	return nil
}

const ApiPostMessage = "/api/chat.postMessage"

func (this *Account) SendMessage(channel, message string) error {
	form := url.Values{}

	if _, ok := this.channels[channel]; !ok {
		return errInvalidChannel
	}

	form.Add("channel", this.channels[channel])
	form.Add("text", message)
	form.Add("ts", fmt.Sprintf("%u.xxxxx2", time.Now().Unix()))
	form.Add("type", "message")
	form.Add("token", this.apiToken)
	form.Add("_x_reason", "webapp_message_send")
	form.Add("_x_mode", "online")
	req, _ := http.NewRequest("POST",
		this.accountCreator.ApiBaseUrl() + ApiPostMessage,
		strings.NewReader(form.Encode()))
	req.Header.Set("Origin", this.accountCreator.ApiBaseUrl())
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := this.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return errInvalidStatusCode
	}

	type Response struct {
		Ok bool `json:"ok"`
	}
	var response Response
	err = json.Unmarshal(b, &response)
	if err != nil {
		return err
	}

	if !response.Ok {
		return errSlackApi
	}

	return nil
}
