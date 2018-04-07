package chos

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

const (
	awsErrCodeNoSuchBucket = "NoSuchBucket"
	awsErrCodeNotFound     = "NotFound"
)

type ChosConfig struct {
	Name      string
	Region    string
	Device    string
	AMI       string
	SpotPrice string
	SSHKey    string
}

func AWSApply(config ChosConfig) {
	log.Info("Checking AWS credentials")
	checkAWSCreds(config.Region)

	log.Infof("Creating S3 bucket %s", config.Name)
	s3BucketSetup(config)

	terraformConf, err := generateTerraformConfig(config)
	if err != nil {
		log.Fatalln("Failed to generate config:", err)
	}

	terraformClient, err := NewTerraformClient(terraformConf, os.Stdout, os.Stdin)
	if err != nil {
		log.Fatalln("Failed to create client:", err)
	}
	defer terraformClient.Cleanup()

	log.Info("Creating AWS resources")
	err = terraformClient.Apply()
	if err != nil {
		log.Fatalln("Failed to create AWS resources:", err)
	}
}

func AWSDestroy(config ChosConfig) {
	log.Info("Checking AWS credentials")
	checkAWSCreds(config.Region)

	terraformConf, err := generateTerraformConfig(config)
	if err != nil {
		log.Fatalln("Failed to generate config:", err)
	}

	terraformClient, err := NewTerraformClient(terraformConf, os.Stdout, os.Stdin)
	if err != nil {
		log.Fatalln("Failed to create client:", err)
	}

	log.Info("Destroying AWS resources")
	err = terraformClient.Destroy()
	if err != nil {
		log.Fatalln("Failed to destroy", err)
	}
	log.Info("Successfully removed AWS resources")
}

func checkAWSCreds(region string) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		log.Fatalln("Failed to create new AWS session:", err)
	}
	s3Client := s3.New(sess, &aws.Config{Region: &region})
	_, err = s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		log.Fatalln("Unable to list S3 buckets - make sure you have valid admin AWS credentials")
	}
}

func s3BucketSetup(config ChosConfig) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		log.Fatalln("Failed to create new AWS session:", err)
	}
	s3Client := s3.New(sess, &aws.Config{Region: &config.Region})

	_, err = s3Client.HeadBucket(&s3.HeadBucketInput{Bucket: &config.Name})
	if err != nil {
		awsErrCode := err.(awserr.Error).Code()
		if awsErrCode != awsErrCodeNotFound && awsErrCode != awsErrCodeNoSuchBucket {
			log.Fatalln("Unknown S3 error code:", err)
		}

		bucketInput := &s3.CreateBucketInput{
			Bucket: &config.Name,
		}
		// NOTE the location constraint should only be set if using a bucket OTHER than us-east-1
		// http://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUT.html
		if config.Region != "us-east-1" {
			bucketInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
				LocationConstraint: &config.Region,
			}
		}

		_, err = s3Client.CreateBucket(bucketInput)
		if err != nil {
			log.Fatalf("Failed to create bucket %s - note that this bucket name must be globally unique. %v", config.Name, err)
		}
	}
}
