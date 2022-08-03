# onesecmail #

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/frei-0xff/onesecmail)

onesecmail is an unofficial Go client for the [1secmail API](https://www.1secmail.com/api/).

## Installation ##

```bash
go get github.com/frei-0xff/onesecmail
```

## Example ##

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/frei-0xff/onesecmail"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
	defer cancel()

	client := onesecmail.Client{}

	mailBoxes, _ := client.GenRandomMailbox(ctx, 10)
	fmt.Println(mailBoxes)

	domains, _ := client.GetDomainList(ctx)
	fmt.Println(domains)

	login := "demo"
	domain := "1secmail.com"
	messages, _ := client.GetMessages(ctx, login, domain)
	for _, m := range messages {
		fmt.Println(m)

		message, _ := client.ReadMessage(ctx, login, domain, m.ID)
		fmt.Println(message)

		for _, a := range message.Attachments {
			attachment, _ := client.DownloadAttachment(ctx, login, domain, m.ID, a.Filename)
			fmt.Println(a, len(attachment))
		}
	}
}
```
