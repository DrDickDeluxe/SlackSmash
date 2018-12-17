package main

import (
	"io/ioutil"
	"time"
	"math/rand"
	"encoding/json"
	"errors"
)

var errInvalidInvite error = errors.New("Invalid invite URL")
var errEmptyProxies error = errors.New("No proxies specified")
var errEmptyMailboxes error = errors.New("No mailboxes specified")
var errEmptyChannels error = errors.New("No channels specified")
var errEmptyMessages error = errors.New("No messages specified")
var errEmptyNames error = errors.New("No names specified")

type Config struct {
	rng				*rand.Rand
	MaxConcurrent	int `json:"max_concurrent"`
	InviteUrl		string `json:"invite_url"`
	Creator			string `json:"account_creator"`
	Proxies			[]string `json:"proxies"`
	Mailboxes		[]string `json:"mailboxes"`
	Channels		[]string `json:"channels"`
	Messages		[]string `json:"messages"`
	DelayPerAccount	int `json:"delay_per_account"`
	DelayPerMessage	int `json:"delay_per_message"`
	Names			[]string `json:"names"`
}

func NewConfig(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	cfg.rng = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))

	if cfg.InviteUrl == "" {
		return nil, errInvalidInvite
	}
	if len(cfg.Proxies) == 0 {
		return nil, errEmptyProxies
	}
	if len(cfg.Mailboxes) == 0 {
		return nil, errEmptyMailboxes
	}
	if len(cfg.Channels) == 0 {
		return nil, errEmptyChannels
	}
	if len(cfg.Messages) == 0 {
		return nil, errEmptyMessages
	}
	if len(cfg.Names) == 0 {
		return nil, errEmptyNames
	}

	return &cfg, nil
}

func (this *Config) AccountCreator() AccountCreatorType {
	return AccountCreatorTypeFromString(this.Creator)
}

func (this *Config) RandomProxy() string {
	return this.Proxies[this.rng.Int() % len(this.Proxies)]
}

func (this *Config) RandomMailbox() MailClientType {
	return MailClientTypeFromString(
		this.Mailboxes[this.rng.Int() % len(this.Mailboxes)])
}

func (this *Config) RandomMessage() string {
	return this.Messages[this.rng.Int() % len(this.Messages)]
}

func (this *Config) RandomName() string {
	return this.Names[this.rng.Int() % len(this.Names)]
}
