package main

import (
	"flag"
	"fmt"
	"time"
	"os"
)

func main() {
	var confFile string
	flag.StringVar(&confFile, "cfg", "", "JSON configuration file")
	flag.Parse()
	if confFile == "" {
		fmt.Println("Usage:", os.Args[0], "-cfg=path/to/config.json")
		return
	}

	var (
		cfg *Config
		err error
	)
	if cfg, err = NewConfig(confFile); err != nil {
		fmt.Println("Unable to load config:", err.Error())
		os.Exit(-1)
	}

	sem := make(chan int, cfg.MaxConcurrent)
	for {
		sem <- 1

		go func() {
			defer func() {
				<-sem
			}()
			client(cfg)
		}()
		time.Sleep(time.Millisecond * time.Duration(cfg.DelayPerAccount))
	}
}

func client(cfg *Config) {
	acc := NewAccount(cfg.InviteUrl, cfg.RandomProxy(),
		cfg.RandomName(), cfg.RandomMailbox(), cfg.AccountCreator())
	if err := acc.Signup(); err != nil {
		fmt.Println("Unable to sign up:", err.Error())
		return
	}
	for {
		for _,c := range cfg.Channels {
			if err := acc.JoinChannel(c); err != nil {
				fmt.Println("Unable to join ", c, ":", err.Error())
				return
			}
		}
		for {
			unableSendMessage := false
			for _,c := range cfg.Channels {
				if err := acc.SendMessage(c, cfg.RandomMessage()); err != nil {
					fmt.Println("Unable to send message:", err.Error())
					unableSendMessage = true
					break
				}

				time.Sleep(time.Millisecond * time.Duration(cfg.DelayPerMessage))
			}
			if unableSendMessage {
				break
			}
		}
	}
}
