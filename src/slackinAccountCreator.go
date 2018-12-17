package main

import (
	"net/http"
)

type SlackinAccountCreator struct {
	mailClient		MailClient
	httpClient		*http.Client
	baseUrl			string
	inviteCode		string
}

func NewSlackinAccountCreator() *SlackinAccountCreator {
	return &SlackinAccountCreator{}
}

func (this *SlackinAccountCreator) ApiBaseUrl() string {
	return this.baseUrl
}

func (this *SlackinAccountCreator) InviteCode() string {
	return this.inviteCode
}
