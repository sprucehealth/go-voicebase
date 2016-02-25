package twilio

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

type MessageService struct {
	client *Client
}

type MessageIFace interface {
	Create(v url.Values) (*Message, *Response, error)
	SendSMS(from, to, body string) (*Message, *Response, error)
	Send(from, to string, params MessageParams) (*Message, *Response, error)
	Get(sid string) (*Message, *Response, error)
	List(params MessageListParams) ([]Message, *Response, error)
	GetMediaMetadata(messageSID, mediaSID string) (*Metadata, *Response, error)
	Delete(sid string) (*Response, error)
	DeleteMedia(mediaURL string) (*Response, error)
}

type Message struct {
	AccountSid  string    `json:"account_sid"`
	ApiVersion  string    `json:"api_version"`
	Body        string    `json:"body"`
	NumSegments int       `json:"num_segments,string"`
	NumMedia    int       `json:"num_media,string"`
	DateCreated Timestamp `json:"date_created,omitempty"`
	DateSent    Timestamp `json:"date_sent,omitempty"`
	DateUpdated Timestamp `json:"date_updated,omitempty"`
	Direction   string    `json:"direction"`
	From        string    `json:"from"`
	Price       Price     `json:"price,omitempty"`
	Sid         string    `json:"sid"`
	Status      string    `json:"status"`
	To          string    `json:"to"`
	Uri         string    `json:"uri"`
}

func (m *Message) IsSent() bool {
	return m.Status == "sent"
}

type MessageParams struct {
	// The text of the message you want to send, limited to 1600 characters.
	Body string

	// The URL of the media you wish to send out with the message. Currently support: gif, png, and jpeg.
	MediaUrl []string

	StatusCallback string
	ApplicationSid string
}

func (p MessageParams) Validates() error {
	if (p.Body == "") && (len(p.MediaUrl) == 0) {
		return errors.New(`One of the "Body" or "MediaUrl" is required.`)
	}

	return nil
}

func (s *MessageService) Create(v url.Values) (*Message, *Response, error) {
	u := s.client.EndPoint("Messages")

	req, _ := s.client.NewRequest("POST", u.String(), strings.NewReader(v.Encode()))

	m := new(Message)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, err
}

// Shortcut for sending SMS with no optional parameters support.
func (s *MessageService) SendSMS(from, to, body string) (*Message, *Response, error) {
	return s.Send(from, to, MessageParams{Body: body})
}

// Send Message with options. It's support required and optional parameters.
//
// One of these parameters is required:
//
//	Body     : The text of the message you want to send
//	MediaUrl : The URL of the media you wish to send out with the message. Currently support: gif, png, and jpeg.
//
// Optional parameters:
//
//	StatusCallback : A URL that Twilio will POST to when your message is processed.
//	ApplicationSid : Twilio will POST `MessageSid` as well as other statuses to the URL in the `MessageStatusCallback` property of this application
func (s *MessageService) Send(from, to string, params MessageParams) (*Message, *Response, error) {
	err := params.Validates()
	if err != nil {
		return nil, nil, err
	}

	v, err := query.Values(&params)
	if err != nil {
		return nil, nil, err
	}

	v.Set("From", from)
	v.Set("To", to)

	return s.Create(v)
}

func (s *MessageService) Get(sid string) (*Message, *Response, error) {
	u := s.client.EndPoint("Messages", sid)

	req, _ := s.client.NewRequest("GET", u.String(), nil)

	m := new(Message)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, err
}

type MessageListParams struct {
	To       string
	From     string
	DateSent string
	PageSize int
}

func (s *MessageService) List(params MessageListParams) ([]Message, *Response, error) {
	u := s.client.EndPoint("Messages")
	v, err := query.Values(&params)
	if err != nil {
		return nil, nil, err
	}

	req, _ := s.client.NewRequest("GET", u.String(), strings.NewReader(v.Encode()))

	// Helper struct for handling the listing
	type list struct {
		Pagination
		Messages []Message `json:"messages"`
	}

	l := new(list)
	resp, err := s.client.Do(req, l)
	if err != nil {
		return nil, resp, err
	}

	resp.Pagination = l.Pagination

	return l.Messages, resp, err
}

// ParseMediaSID expects the url to be of the form /2010-04-01/Accounts/{AccountSid}/Messages/{MessageSid}/Media/{MediaSid}
// to be able to parse out the mediaID.
func ParseMediaSID(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	parts := strings.Split(parsedURL.Path, "/")
	if len(parts) != 8 {
		return "", fmt.Errorf("Expected URI of the form /2010-04-01/Accounts/{AccountSid}/Messages/{MessageSid}/Media/{MediaSid}, but got %s", parsedURL.Path)
	}

	return strings.Split(parts[7], ".")[0], nil
}

func (s *MessageService) GetMediaMetadata(messageSID, mediaSID string) (*Metadata, *Response, error) {
	u := s.client.EndPoint("Messages", messageSID, "Media", mediaSID)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	met := new(Metadata)
	res, err := s.client.Do(req, met)
	if err != nil {
		return nil, nil, err
	}
	return met, res, nil
}

func (s *MessageService) Delete(sid string) (*Response, error) {
	u := s.client.EndPoint("Messages", sid)

	req, err := s.client.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *MessageService) DeleteMedia(url string) (*Response, error) {

	req, err := s.client.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
