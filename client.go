package onesecmail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
)

// DefaultBaseURL is where 1secmail expects API calls.
const DefaultBaseURL = "https://www.1secmail.com/api/v1/"

// Version is the current library's version: sent with User-Agent
const Version = "0.1"

// Client interacts with the 1secmail API
type Client struct {
	HTTPClient http.Client
	BaseURL    string
	VerboseLog Log
}

// Log is a function that Client can take to optionally verbose log what it does internally
type Log func(...interface{})

// ErrNotExpectedJSON is returned by API calls when the response isn't expected JSON
type ErrNotExpectedJSON struct {
	OriginalBody string
	Err          error
}

func (e *ErrNotExpectedJSON) Error() string {
	return fmt.Sprintf("Unexpected JSON: %s from %s", e.Err.Error(), e.OriginalBody)
}

// ErrBadStatusCode is returned when the API returns a non 200 error code
type ErrBadStatusCode struct {
	OriginalBody string
	Code         int
}

// Structure of an email message from emails list
type MessageItem struct {
	ID      int    `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// Structure of an email message
type Message struct {
	ID          int    `json:"id"`
	From        string `json:"from"`
	Subject     string `json:"subject"`
	Date        string `json:"date"`
	Attachments []struct {
		ContentType string `json:"contentType"`
		Filename    string `json:"filename"`
		Size        int    `json:"size"`
	} `json:"attachments"`
	Body     string `json:"body"`
	TextBody string `json:"textBody"`
	HTMLBody string `json:"htmlBody"`
}

func (e *ErrBadStatusCode) Error() string {
	return fmt.Sprintf("Invalid status code: %d", e.Code)
}

func (c *Client) verboseLog(v ...interface{}) {
	if c.VerboseLog != nil {
		c.VerboseLog(v...)
	}
}

func (c *Client) doReqURL(ctx context.Context, u string, jsonInto interface{}, getRawData bool) error {
	c.verboseLog("fetching", u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("github.com/frei-0xff/onesecmail/%s (gover %s)", Version, runtime.Version()))
	req = req.WithContext(ctx)
	client := &c.HTTPClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var b bytes.Buffer
	if _, err := io.Copy(&b, resp.Body); err != nil {
		return err
	}
	if getRawData {
		rawData := jsonInto.(*[]byte)
		*rawData = b.Bytes()
		return nil
	}
	debug := b.String()
	if resp.StatusCode != http.StatusOK {
		c.verboseLog("Invalid status code", resp.StatusCode)
		return &ErrBadStatusCode{
			OriginalBody: debug,
			Code:         resp.StatusCode,
		}
	}
	c.verboseLog("Fetch result", debug)
	if err := json.NewDecoder(&b).Decode(jsonInto); err != nil {
		return &ErrNotExpectedJSON{
			OriginalBody: debug,
			Err:          err,
		}
	}
	return nil
}

// Generate <count> random email addresses
func (c *Client) GenRandomMailbox(ctx context.Context, count int) ([]string, error) {
	var v []string
	params := url.Values{}
	params.Set("action", "genRandomMailbox")
	params.Set("count", strconv.Itoa(count))
	if err := c.doReqURL(ctx, c.url(&params), &v, false); err != nil {
		return nil, err
	}
	return v, nil
}

// Get list of currently active domains on which our system is handling incoming emails at the moment
func (c *Client) GetDomainList(ctx context.Context) ([]string, error) {
	var v []string
	params := url.Values{}
	params.Set("action", "getDomainList")
	if err := c.doReqURL(ctx, c.url(&params), &v, false); err != nil {
		return nil, err
	}
	return v, nil
}

// Check and get a list of emails for a mailbox <login>@<domain>
func (c *Client) GetMessages(ctx context.Context, login string, domain string) ([]MessageItem, error) {
	var v []MessageItem
	params := url.Values{}
	params.Set("action", "getMessages")
	params.Set("login", login)
	params.Set("domain", domain)
	if err := c.doReqURL(ctx, c.url(&params), &v, false); err != nil {
		return nil, err
	}
	return v, nil
}

// Get message <id> from a mailbox <login>@<domain>
func (c *Client) ReadMessage(ctx context.Context, login string, domain string, id int) (*Message, error) {
	var v Message
	params := url.Values{}
	params.Set("action", "readMessage")
	params.Set("login", login)
	params.Set("domain", domain)
	params.Set("id", strconv.Itoa(id))
	if err := c.doReqURL(ctx, c.url(&params), &v, false); err != nil {
		return nil, err
	}
	return &v, nil
}

// Download attachment <filename> of message <id> from a mailbox <login>@<domain>
func (c *Client) DownloadAttachment(ctx context.Context, login string, domain string, id int, filename string) ([]byte, error) {
	var v []byte
	params := url.Values{}
	params.Set("action", "download")
	params.Set("login", login)
	params.Set("domain", domain)
	params.Set("id", strconv.Itoa(id))
	params.Set("file", filename)
	if err := c.doReqURL(ctx, c.url(&params), &v, true); err != nil {
		return nil, err
	}
	return v, nil
}

func (c *Client) urlBase() string {
	base := c.BaseURL
	if c.BaseURL == "" {
		base = DefaultBaseURL
	}
	return base
}

func (c *Client) url(params *url.Values) string {
	return fmt.Sprintf("%s?%s", c.urlBase(), params.Encode())
}
