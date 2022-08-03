package main

import (
	"context"
	"fmt"
	"time"

	"github.com/frei-0xff/onesecmail"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
	defer cancel()
	client := onesecmail.Client{}
	mailBoxes, err := client.GenRandomMailbox(ctx, 10)
	check(err)
	fmt.Println(mailBoxes)
	domains, err := client.GetDomainList(ctx)
	check(err)
	fmt.Println(domains)
	login := "demo"
	domain := "1secmail.com"
	messages, err := client.GetMessages(ctx, login, domain)
	check(err)
	for _, m := range messages {
		fmt.Println(m)
		message, err := client.ReadMessage(ctx, login, domain, m.ID)
		check(err)
		fmt.Println(message)
		for _, a := range message.Attachments {
			attachment, err := client.DownloadAttachment(ctx, login, domain, m.ID, a.Filename)
			check(err)
			fmt.Println(a, len(attachment))
		}
	}
}
