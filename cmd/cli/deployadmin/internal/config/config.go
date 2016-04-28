package config

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/sprucehealth/backend/boot"
	"github.com/sprucehealth/backend/svc/deploy"
	"google.golang.org/grpc"
)

type Config struct {
	App        *boot.App
	DeployAddr string
}

func (c *Config) DeployClient() (deploy.DeployClient, error) {
	if c.DeployAddr == "" {
		return nil, errors.New("Deploy service address required")
	}
	conn, err := grpc.Dial(c.DeployAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to auth service: %s", err)
	}
	return deploy.NewDeployClient(conn), nil
}

func (c *Config) STSClient() (stsiface.STSAPI, error) {
	awsSession, err := c.App.AWSSession()
	if err != nil {
		return nil, fmt.Errorf("Unable to create AWSSession: %s", err)
	}
	return sts.New(awsSession), nil
}