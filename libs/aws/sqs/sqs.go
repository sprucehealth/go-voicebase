package sqs

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sprucehealth/backend/libs/aws"
)

const version = "2012-11-05"

const (
	DefaultVisibility = -1
	DefaultWaitTime   = -1
)

type SQS struct {
	aws.Region
	Client *aws.Client
	Debug  bool

	host string
}

func QueueName(url string) string {
	idx := strings.LastIndex(url, "/")
	if idx < 0 {
		return ""
	}
	return url[idx+1:]
}

func (sqs *SQS) Request(endpoint, action string, args url.Values, response interface{}) error {
	if args == nil {
		args = url.Values{}
	}
	if args.Get("Version") == "" {
		args.Set("Version", version)
	}
	if args.Get("Timestamp") == "" && args.Get("Expires") == "" {
		args.Set("Timestamp", time.Now().In(time.UTC).Format(time.RFC3339))
	}
	args.Set("Action", action)
	if endpoint == "" {
		endpoint = sqs.SQSEndpoint
	}
	res, err := sqs.Client.PostForm(endpoint, args)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return ParseErrorResponse(res)
	}
	var body io.Reader = res.Body
	if sqs.Debug {
		body = io.TeeReader(body, os.Stderr)
	}
	dec := xml.NewDecoder(body)
	return dec.Decode(response)
}

func (sqs *SQS) DeleteMessage(queueUrl, receiptHandle string) error {
	args := url.Values{
		"ReceiptHandle": []string{receiptHandle},
	}
	res := simpleResponse{}
	if err := sqs.Request(queueUrl, "DeleteMessage", args, &res); err != nil {
		return err
	}
	return nil
}

func (sqs *SQS) GetQueueUrl(queueName, queueOwnerAWSAccountId string) (string, error) {
	args := url.Values{
		"QueueName": []string{queueName},
	}
	if queueOwnerAWSAccountId != "" {
		args.Set("QueueOwnerAWSAccountId", queueOwnerAWSAccountId)
	}
	res := getQueueUrlResponse{}
	if err := sqs.Request("", "GetQueueUrl", args, &res); err != nil {
		return "", err
	}
	return res.Url, nil
}

func (sqs *SQS) ListQueues(namePrefix string) ([]string, error) {
	args := url.Values{}
	if namePrefix != "" {
		args.Set("QueueNamePrefix", namePrefix)
	}
	res := listQueuesResponse{}
	if err := sqs.Request("", "ListQueues", args, &res); err != nil {
		return nil, err
	}
	return res.QueueUrls, nil
}

func (sqs *SQS) SendMessage(queueUrl string, delaySeconds int, messageBody string) error {
	args := url.Values{}

	if delaySeconds > 0 {
		args.Set("DelaySeconds", strconv.Itoa(delaySeconds))
	}

	args.Set("MessageBody", messageBody)
	res := sendMessageResponse{}
	return sqs.Request(queueUrl, "SendMessage", args, &res)
}

/*
attributes: A list of attributes that need to be returned along with each message.
maxNumberOfMessages: The maximum number of messages to return. Amazon SQS never returns more messages than this value but may return fewer.
queueUrl: The URL of the Amazon SQS queue to take action on.
visibilityTimeout: The duration (in seconds) that the received messages are hidden from subsequent retrieve requests after being retrieved by a ReceiveMessage request.
waitTimeSeconds: The duration (in seconds) for which the call will wait for a message to arrive in the queue before returning. If a message is available, the call will return sooner than WaitTimeSeconds.
*/

func (sqs *SQS) ReceiveMessage(queueUrl string, attributes []AttributeName, maxNumberOfMessages, visibilityTimeout, waitTimeSeconds int) ([]*Message, error) {
	args := url.Values{}
	for i, attr := range attributes {
		args.Set(fmt.Sprintf("AttributeName.%d", i+1), string(attr))
	}
	if maxNumberOfMessages > 0 {
		args.Set("MaxNumberOfMessages", strconv.Itoa(maxNumberOfMessages))
	}
	if visibilityTimeout >= 0 {
		args.Set("VisibilityTimeout", strconv.Itoa(visibilityTimeout))
	}
	if waitTimeSeconds >= 0 {
		args.Set("WaitTimeSeconds", strconv.Itoa(waitTimeSeconds))
	}
	res := receiveMessageResponse{}
	if err := sqs.Request(queueUrl, "ReceiveMessage", args, &res); err != nil {
		return nil, err
	}
	return res.Messages, nil
}
