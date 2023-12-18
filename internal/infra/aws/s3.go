package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Api interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
