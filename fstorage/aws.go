package fstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

type AWSS3Credentials struct {
	AccessKey string
	SecretKey string
	Region    string
	URL       string
}

func GetS3Client(ctx context.Context, creds AWSS3Credentials) (*s3.Client, error) {
	staticProvider := credentials.NewStaticCredentialsProvider(
		creds.AccessKey,
		creds.SecretKey,
		"",
	)

	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(creds.Region),
		awsConfig.WithCredentialsProvider(staticProvider),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(creds.URL)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationUnset
	})

	return s3Client, nil
}

type AWSS3FileStorage struct {
	Client *s3.Client
}

func NewAWSS3FileStorage(ctx context.Context, creds AWSS3Credentials) (*AWSS3FileStorage, error) {
	client, err := GetS3Client(ctx, creds)
	if err != nil {
		return nil, err
	}

	return &AWSS3FileStorage{Client: client}, nil
}

func (s AWSS3FileStorage) UploadFile(
	ctx context.Context,
	bucket string,
	filepath string,
	content *bytes.Buffer,
	metadata map[string]string,
) error {
	tags := url.Values{}
	for k, v := range metadata {
		tags.Add(k, v)
	}

	request := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(filepath),
		Body:          content,
		ContentLength: aws.Int64(int64(content.Len())),
		Tagging:       aws.String(tags.Encode()),
	}

	_, err := s.Client.PutObject(
		ctx,
		request,
		s3.WithAPIOptions(
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to upload file '%s', err: %w", filepath, err)
	}

	return nil
}

// Generates a temporary download URL for the provided bucket and filepath
func (s AWSS3FileStorage) GetDownloadPresignedURL(
	ctx context.Context,
	bucket string,
	filepath string,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)
	obj, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filepath),
		},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		return "", err
	}

	return obj.URL, nil
}

// Generates a temporary upload URL for the provided bucket and filepath
func (s AWSS3FileStorage) GetUploadPresignedURL(
	ctx context.Context,
	bucket string,
	filepath string,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)
	obj, err := presignClient.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filepath),
		},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		return "", err
	}

	return obj.URL, nil
}

func (s AWSS3FileStorage) BucketExists(
	ctx context.Context,
	bucket string,
) (bool, error) {
	_, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				exists = false
				err = nil
			default:
				return false, err
			}
		}
	}

	return exists, err
}

func (s AWSS3FileStorage) CreateBucket(
	ctx context.Context,
	region string,
	bucket string,
) error {
	req := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	if region != "us-east-1" {
		req.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	_, err := s.Client.CreateBucket(ctx, req)
	if err != nil {
		var owned *types.BucketAlreadyOwnedByYou
		var exists *types.BucketAlreadyExists

		if errors.As(err, &owned) {
			err = owned
		} else if errors.As(err, &exists) {
			err = exists
		}

		return err
	}

	return nil
}
