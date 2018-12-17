package main

import (
	"fmt"
	"net/http"
	"regexp"
	"time"
	"strings"
	"errors"
)

var errInviteLink error = errors.New("Slack invite did not have link")
var errInviteTimeout error = errors.New("Timed out waiting for Slack email")

type Mail interface {
	Subject() string
	Body() string
}

type MailClient interface {
	Signup() error
	Address() string
	Inbox() ([]Mail, error)
}

const (
	UnknownMail		= iota
	TenMinuteMail	= iota
	GuerrillaMail	= iota
)
type MailClientType int

func MailClientTypeFromString(name string) MailClientType {
	clients := map[string]MailClientType {
		"10MinuteMail": TenMinuteMail,
		"GuerrillaMail": GuerrillaMail,
	}
	if t, ok := clients[name]; ok {
		return t
	}
	return UnknownMail
}

func NewMailClient(t MailClientType, httpClient *http.Client) MailClient {
	switch t {
	case TenMinuteMail:
		return NewTmmMailClient(httpClient)
	case GuerrillaMail:
		return NewGuerrillaMailClient(httpClient)
	default:
		return nil
	}
}

var reSlackUrl *regexp.Regexp = regexp.MustCompile(
	`join\.slack\.com\/t\/([\w\d]*)\/invite\/(.*)\?`)

func PollInboxForSlackMessage(m MailClient, to time.Duration) (
		string, string, error) {
	for start := time.Now(); time.Now().Sub(start) < to; {
		messages, err := m.Inbox()
		if err != nil {
			return "", "", err
		}
		for _,msg := range messages {
			if !strings.Contains(msg.Subject(), "has invited you to join") {
				continue
			}

			capture := reSlackUrl.FindStringSubmatch(msg.Body())
			if len(capture) < 3 {
				return "", "", errInviteLink
			}

			baseUrl := fmt.Sprintf("https://%s.slack.com", capture[1])
			inviteCode := capture[2]
			return baseUrl, inviteCode, nil
		}
		time.Sleep(time.Second * 5)
	}
	return "", "", errInviteTimeout
}
