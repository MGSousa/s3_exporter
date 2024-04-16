package exporter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/prometheus/common/log"
)

type AWSClient struct {
	s3 struct {
		s3iface.S3API

		bucketsInputs s3.ListBucketsInput
		objectsInputs s3.ListObjectsV2Input
	}
}

func NewAwsSession(url string, ssl, forcePathStyle bool) *AWSClient {
	sess, err := session.NewSession()
	if err != nil {
		log.Errorln("Error creating sessions ", err)
	}

	cfg := aws.NewConfig()
	if url != "" {
		cfg.WithEndpoint(url)
	}

	cfg.WithDisableSSL(ssl)
	cfg.WithS3ForcePathStyle(forcePathStyle)

	// TODO: change these settings on final
	cfg.WithRegion("us-east-2")
	cfg.WithCredentials(credentials.NewStaticCredentials("test", "test", ""))

	return &AWSClient{
		s3: struct {
			s3iface.S3API
			bucketsInputs s3.ListBucketsInput
			objectsInputs s3.ListObjectsV2Input
		}{
			S3API: s3.New(sess, cfg),
		},
	}
}

func (a *AWSClient) StringValue(s *string) string {
	return aws.StringValue(s)
}

func (a *AWSClient) ToString(s string) *string {
	return aws.String(s)
}
