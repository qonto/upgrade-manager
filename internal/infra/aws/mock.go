package aws

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
)

type S3Mock struct {
	mock.Mock
}

func (m *S3Mock) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(*params.Bucket)
	b := args.Get(0).([]byte) //nolint
	output := &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(b))}
	return output, args.Error(1) //nolint
}
