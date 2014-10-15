package s3

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

type ErrBadStatusCode int

func (e ErrBadStatusCode) Error() string {
	return fmt.Sprintf("bad status code %d", int(e))
}

func dumpResponse(res *http.Response) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, res.Body); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf.String())
}

type ErrorResponse struct {
	Code       string `xml:"Code"`
	Message    string `xml:"Message"`
	RequestID  string `xml:"RequestId"`
	ContentMD5 string `xml:"Content-MD5"`
	HostID     string `xml:"HostId"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", er.Code, er.Message)
}

func ParseErrorResponse(res *http.Response) error {
	dec := xml.NewDecoder(res.Body)
	var er ErrorResponse
	if err := dec.Decode(&er); err != nil {
		return err
	}
	return &er
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	switch err {
	case io.ErrUnexpectedEOF, io.EOF:
		return true
	}
	switch e := err.(type) {
	case *net.DNSError:
		return true
	case *net.OpError:
		switch e.Op {
		case "read", "write":
			return true
		}
	case *ErrorResponse:
		switch e.Code {
		case "InternalError", "NoSuchUpload", "NoSuchBucket":
			return true
		}
	}
	return false
}
