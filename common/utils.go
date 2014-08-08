package common

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"

	"github.com/sprucehealth/backend/libs/aws"
	"github.com/sprucehealth/backend/libs/aws/sqs"
	goamz "github.com/sprucehealth/backend/third_party/launchpad.net/goamz/aws"
)

// Any structure that implements the Typed interface
// requires a string that defines the type of the structure
type Typed interface {
	TypeName() string
}

type TypedData struct {
	Data []byte
	Type string
}

type ClientView interface {
	Validate() error
}

func (t *TypedData) TypeName() string {
	return t.Type
}

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

type ERxSourceType int64

const (
	ERxType ERxSourceType = iota
	RefillRxType
	UnlinkedDNTFTreatmentType
)

type PrescriptionStatusCheckMessage struct {
	PatientId      int64
	DoctorId       int64
	EventCheckType ERxSourceType
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

func SeekerSize(sk io.Seeker) (int64, error) {
	size, err := sk.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}
	_, err = sk.Seek(0, os.SEEK_SET)
	return size, err
}
