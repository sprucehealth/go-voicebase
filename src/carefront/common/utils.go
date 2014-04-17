package common

import (
	"carefront/libs/aws"
	"carefront/libs/aws/sqs"
	"crypto/rand"
	"encoding/base64"

	goamz "launchpad.net/goamz/aws"
)

func GenerateToken() (string, error) {
	tokBytes := make([]byte, 16)
	if _, err := rand.Read(tokBytes); err != nil {
		return "", err
	}

	tok := base64.URLEncoding.EncodeToString(tokBytes)
	return tok, nil
}

func AWSAuthAdapter(auth aws.Auth) goamz.Auth {
	keys := auth.Keys()
	return goamz.Auth{
		AccessKey: keys.AccessKey,
		SecretKey: keys.SecretKey,
		Token:     keys.Token,
	}
}

type StatusEventCheckType int64

const (
	ERxType StatusEventCheckType = iota
	RefillRxType
	UnlinkedDNTFTreatmentType
)

type PrescriptionStatusCheckMessage struct {
	PatientId      int64
	DoctorId       int64
	EventCheckType StatusEventCheckType
}

type SQSQueue struct {
	QueueService sqs.SQSService
	QueueUrl     string
}

func NewQueue(auth aws.Auth, region aws.Region, queueName string) (*SQSQueue, error) {
	awsClient := &aws.Client{
		Auth: auth,
	}

	sq := &sqs.SQS{
		Region: region,
		Client: awsClient,
	}

	queueUrl, err := sq.GetQueueUrl(queueName, "")
	if err != nil {
		return nil, err
	}

	return &SQSQueue{
		QueueService: sq,
		QueueUrl:     queueUrl,
	}, nil
}
