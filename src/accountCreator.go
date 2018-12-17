package main

import "net/http"

type AccountCreator interface {
	Create(name string) error
	ApiBaseUrl() string
	InviteCode() string
}

const (
	UnkonwnCreator	= iota
	PublicSlack		= iota
	Slackin			= iota
	LaraChat		= iota
)
type AccountCreatorType int

func NewAccountCreator(t AccountCreatorType, inviteUrl string,
		mc MailClient, hc *http.Client) AccountCreator {
	switch t {
	case PublicSlack:
		return NewPublicSlackAccountCreator(inviteUrl, mc, hc)
	case Slackin:
		return nil
	case LaraChat:
		return NewLaraChatAccountCreator(inviteUrl, mc, hc)
	default:
		return nil
	}
}

func AccountCreatorTypeFromString(name string) AccountCreatorType {
	creators := map[string]AccountCreatorType {
		"PublicSlack": PublicSlack,
		"Slackin": Slackin,
		"LaraChat": LaraChat,
	}
	if t, ok := creators[name]; ok {
		return t
	}
	return UnkonwnCreator
}
