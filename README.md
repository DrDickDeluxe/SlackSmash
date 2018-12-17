# Slack Smasher

Smashing The Slack For Fun And Profit.

Do you have too much time on your hands? Do you have a malicious streak inside
of you? The Slack Smasher is probably not what you're looking for!

Seriously, there are much better ways of having fun.

## Installing

Slack Smasher depends on `golang.org/x/net/proxy`. The included `Makefile`
should automatically fetch it for you, but if it fails, here is how to do it
manually:

- Ensure that `GOPATH` is set to a folder. `export GOPATH=$HOME/Go`
- Run `go get golang.org/x/net/proxy`

A simple `make` should suffice, outputting the compiled binary to `bin/ss`.

## Config

An example configuration has been placed in the project root called
`ExampleConf.json`.

A more in-depth explanation of available configuration items is available here:

| Name              | Type                | Description                                                             | Example                                           |
|-------------------|---------------------|-------------------------------------------------------------------------|---------------------------------------------------|
| invite_url        | String/URL          | URL for invite. Depends on account_creator type                         | https://publicslack.com/slacks/random/invites/new |
| account_creator   | String/Enum         | Process required to receive a Slack invite via email                    | PublicSlack, Slackin, LaraChat                    |
| max_concurrent    | Int/Positive        | Maximum number of concurrent accounts active speaking                   | 128                                               |
| proxies           | String List/IP:Port | List of SOCKS5 proxies. Do not use the full SOCKS5 URL specification    | 127.0.0.1:9050                                    |
| mailboxes         | String List/Enum    | Disposable email services to receive invites                            | 10MinuteMail, GuerrillaMail                       |
| channels          | String List/Channel | List of channels to join. Each account will join each channel           | general, discussions                              |
| messages          | String List/Message | List of messages to send. Each bot will keep randomly picking a message | Hello world, Goodbye world                        |
| delay_per_account | Int/Milliseconds    | How long to wait before creating a new account                          | 1000                                              |
| delay_per_message | Int/Millisecond     | How long each bot should wait in between sending messages               | 4000                                              |
| names             | String List/Name    | List of display names to sign up with                                   | Foo, Bar, Baz                                     |

## Usage

Run `bin/ss -cfg "path/to/config.json"`. If there are any errors or `cfg` is not
specified, it will fail to launch.

## Extending

Without modifying any core features, it is possible to easily add support for
additional Disposable Email services and/or Account Creation processes. If you
are a shitty programmer and/or a skid you may skip the rest of this document.
If you don't know which category you fall in you are probably a shitlord and
can skip the rest of this document.

### Adding a Account Creation process

Account Creation services should implement the functions in the interface
`AccountCreator`, documented in `accountCreator.go:5`. At a minimum, each
service must implement the following functions:

- Create(name string) error
- ApiBaseUrl() string
- InviteCode() string

There exist two finished reference implementations:

- publicSlackAccountCreator.go
- laraChatAccountCreator.go

_Create(name string) error_ accepts the display name, and is responsible for
performing the necessary actions to get an invite code emailed to itself, and
to extract data from that link for the below two functions.

_ApiBaseUrl() string_ returns the Slack base URL, such as
`https://myteam.slack.com`.

_InviteCode() string_ returns the raw invite code from the URL. This code is
used to actually create the Slack Account.

Lastly, to add support to loading the Account Creator from the config, you
must modify `accountCreator.go` in three places.

- _accountCreator.go:11_ - Add an "Enum" that represents your creator
- _accountCreator.go:22_ - Add your Constructor to the switch statement
- _accountCreator.go:34_ - Add a mapping to convert a string to "Enum"

If you wish to contribute back, a fun place to start might be adding in support
for ReCaptcha-based processes :-). BestCaptchaSolver charges 2,5USD for 1k solves,
though if you have anything against indentured servitude at a Chinese firm you
should probably look somewhere else.

### Adding a Disposable Email service

Disposable Email services should implement the functions in the interface
`Mail` and `MailClient`, documented in `mailClient.go:20` and `mailClient.go:20`
respectively. At a minimum, each service must implement the following
functions:

- Mail.Subject() string
- Mail.Body() string
- MailClient.Signup() error
- MailClient.Address() string
- MailClient.Inbox() ([]Mail, error)

There exist two finished reference implementations:

- tmmMailClient.go
- guerrillaMailClient.go

_Mail.Subject() string_ returns the subject of a given email

_Mail.Body() string_ returns the body of a given email. It does not particularly
matter if this includes HTML markup or not. If a plain text version is available
it is preferred, but no extra effort should be taken.

_MailClient.Signup() error_ Signs up and receives a unique email address

_MailClient.Address() string_ returns the unique email address

_MailClient.Inbox() ([]Mail, error)_ Attempts to access the service's inbox and
returns an array of emails.

Lastly, to add support to using this Disposable Email service from the config,
you must modify `mailClient.go` in three places.

- _mailClient.go:26_ - Add an "Enum" that represents your creator
- _mailClient.go:34_ - Add a mapping to convert a string to "Enum"
- _mailClient.go:44_ - Add your Constructor to the switch statement

## Donations

lmao fam donate to yourself you fucking poorboy liberal cuck

## Enough talk, get to smashing, shoutout to @/jchung

Enjoy!
