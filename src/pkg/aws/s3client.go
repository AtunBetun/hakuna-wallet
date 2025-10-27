package aws

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Client is a minimal wrapper for uploading and generating presigned URLs.
type S3Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
}

// NewS3Client loads AWS configuration and initializes the client.
func NewS3Client(bucket string, cfg *aws.Config) (*S3Client, error) {
	s3Client := s3.NewFromConfig(*cfg)
	presignClient := s3.NewPresignClient(s3Client)
	return &S3Client{
		client:        s3Client,
		presignClient: presignClient,
	}, nil
}

// PresignURL generates a presigned URL for an existing S3 object.
//
// duration defines how long the URL will remain valid.
// Example: 7 days = 7*24*time.Hour.
func (c *S3Client) PresignURLDefault(ctx context.Context, bucket string, key string) (string, error) {
	return c.PresignURL(ctx, bucket, key, 7*24*time.Hour)
}

func (c *S3Client) PresignURL(ctx context.Context, bucket string, key string, duration time.Duration) (string, error) {
	req, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(duration))
	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}

	return req.URL, nil
}

// UploadFile uploads a local file to S3.
func (c *S3Client) UploadFile(ctx context.Context, bucket, key, filePath string) (*s3.PutObjectOutput, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	out, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return out, nil
}

// UploadBytes uploads an in-memory byte slice to S3.
func (c *S3Client) UploadBytes(ctx context.Context, bucket, key string, data []byte, contentType string) (*s3.PutObjectOutput, error) {
	out, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload bytes to S3: %w", err)
	}
	return out, nil
}
