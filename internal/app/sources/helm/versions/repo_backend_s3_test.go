package versions

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
)

func TestS3RepoBackendGetFile(t *testing.T) {
	testCases := map[string]any{
		"s3://foo-bucket/key/to/bar": true, // url, expected result of test
		"s3://foo-bucket/bar":        true,
		"https://foo-website.com":    false,
	}
	for tc, expectSuccess := range testCases {
		u, err := url.Parse(tc)
		if err != nil {
			t.Errorf("Failed parsing input url, %s", err)
		}
		s3backend := S3HelmRepoBackend{
			s3client: aws.MockS3GetApi(func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				if params.Bucket == nil {
					t.Errorf("No bucket name provided for s3 bucket")
				}
				if params.Key == nil {
					t.Errorf("No key provided for file in s3 bucket")
				}

				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("hello from s3")))}, nil
			}),
		}
		_, err = s3backend.getFile(context.TODO(), u)
		if err != nil {
			if expectSuccess == true {
				t.Errorf("Unexpected test result, expected successful test but the test failed")
			}
		} else {
			if expectSuccess == false {
				t.Errorf("Unexpected test result, expected failed test but the test was successful")
			}
		}
	}
}

func TestS3RepoBackendGetIndexFile(t *testing.T) {
	sampleIndexFile := "../../../_testdata/index.yaml"
	testCases := map[string]any{
		"s3://foo-bucket/key/to/bar": true,
		"s3://foo-bucket/bar":        true,
		"https://foo-website.com":    false,
	}
	for tc, expectSuccess := range testCases {
		u, err := url.Parse(tc)
		if err != nil {
			t.Errorf("Failed parsing input url, %s", err)
		}
		s3backend := S3HelmRepoBackend{
			s3client: aws.MockS3GetApi(func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				if params.Bucket == nil {
					t.Errorf("No bucket name provided for s3 bucket")
				}
				if params.Key == nil {
					t.Errorf("No key provided for file in s3 bucket")
				}
				file, err := os.ReadFile(sampleIndexFile)
				if err != nil {
					t.Errorf("error reading sample file index.yaml at %s", sampleIndexFile)
				}
				return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(file))}, nil
			}),
			bucketUrl: u,
		}
		_, err = s3backend.getIndexFile()
		if err != nil {
			if expectSuccess == true {
				t.Errorf("Unexpected test result, expected successful test but the test failed")
			}
		} else {
			if expectSuccess == false {
				t.Errorf("Unexpected test result, expected failed test but the test was successful")
			}
		}
	}
}

func TestByteSliceToIndexFile(t *testing.T) {
	sampleIndexFile := "../../../_testdata/index.yaml"
	f, err := os.ReadFile(sampleIndexFile)
	if err != nil {
		t.Error(err)
	}
	_, err = byteSliceToIndexFile(f)
	if err != nil {
		t.Error(err)
	}
}
