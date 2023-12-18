package versions

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/qonto/upgrade-manager/internal/infra/aws"
	"helm.sh/helm/v3/pkg/repo"
)

type S3HelmRepoBackend struct {
	s3client  aws.S3Api
	bucketUrl *url.URL
}

// Retrieve a file from an S3 backend based on its full url (s3 bucketname + path)
func (h *S3HelmRepoBackend) getFile(ctx context.Context, url *url.URL) (io.ReadCloser, error) {
	if url.Scheme != "s3" {
		return nil, fmt.Errorf("wrong scheme to get file from s3 url %s, , expected: %s, got: %s", "s3", url, url.Scheme)
	}
	var bucketKey string
	if url.Path != "" && url.Path[0:1] == "/" {
		bucketKey = url.Path[1:]
	}
	f, err := h.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &url.Host,
		Key:    &bucketKey,
	},
	)
	if err != nil {
		return nil, err
	}
	return f.Body, nil
}

// Retrieve the index.yaml file from the helm chart repository and return them
// in chronological order
func (h *S3HelmRepoBackend) getIndexFile() (*repo.IndexFile, error) {
	var index *repo.IndexFile
	indexUrl, err := url.Parse(h.bucketUrl.String() + "/index.yaml")
	if err != nil {
		return index, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	body, err := h.getFile(ctx, indexUrl)
	if err != nil {
		return index, err
	}
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	index, err = byteSliceToIndexFile(raw)
	if err != nil {
		return index, err
	}
	index.SortEntries()
	if err := body.Close(); err != nil {
		return index, err
	}
	return index, nil
}

func byteSliceToIndexFile(data []byte) (*repo.IndexFile, error) {
	f, err := os.CreateTemp(os.TempDir(), "*.yaml")
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(f.Name(), data, 0o600)
	if err != nil {
		return nil, err
	}
	idx, err := repo.LoadIndexFile(f.Name())
	if err != nil {
		return nil, err
	}
	err = os.Remove(f.Name())
	if err != nil {
		return nil, err
	}
	return idx, err
}
