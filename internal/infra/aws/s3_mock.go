package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MockS3GetApi func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)

func (m MockS3GetApi) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m(ctx, params, optFns...)
}
