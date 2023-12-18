package versions

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"helm.sh/helm/v3/pkg/repo"
)

type HelmRepoBackend struct {
	ChartRepo *repo.ChartRepository
}

// get index file from helm chart repository and return it with
// sorted chart versions
func (h *HelmRepoBackend) getIndexFile() (*repo.IndexFile, error) {
	indexFilePath, err := h.ChartRepo.DownloadIndexFile()
	if err != nil {
		return &repo.IndexFile{}, err
	}
	index, err := repo.LoadIndexFile(indexFilePath)
	if err != nil {
		return &repo.IndexFile{}, err
	}
	index.SortEntries()

	return index, nil
}

// get file from remote helm chart repository
// TODO: configure http server
func (h *HelmRepoBackend) getFile(ctx context.Context, url *url.URL) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
