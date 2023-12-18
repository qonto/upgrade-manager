package registry

import (
	"fmt"

	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Auth struct {
	AWS bool `yaml:"aws"`
}

type Config struct {
	EnableDateRetrieval bool `yaml:"enable-date-retrieval"`
	Auth                Auth `yaml:"auth"`
}

type Client struct {
	config  *Config
	options []remote.Option
}

func New(config *Config) (*Client, error) {
	options := []remote.Option{}
	if config.Auth.AWS {
		ecrHelper := ecr.NewECRHelper(ecr.WithClientFactory(api.DefaultClientFactory{}))
		auth := remote.WithAuthFromKeychain(authn.NewKeychainFromHelper(ecrHelper))
		options = append(options, auth)
	}
	return &Client{
		config:  config,
		options: options,
	}, nil
}

func (c *Client) Tags(repositoryString string) ([]string, error) {
	repository, err := name.NewRepository(repositoryString)
	if err != nil {
		return nil, err
	}
	tags, err := remote.List(repository, c.options...)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *Client) ConfigFile(image string) (*v1.ConfigFile, error) {
	if c.config.EnableDateRetrieval {
		ref, err := name.ParseReference(image)
		if err != nil {
			return nil, err
		}
		remote, err := remote.Image(ref, c.options...)
		if err != nil {
			return nil, err
		}
		return remote.ConfigFile()
	}
	return nil, fmt.Errorf("Date retrieval is disabled for this registry (image %s)", image)
}

func (c *Client) ReleaseDateRetrievalEnabled() bool {
	return c.config.EnableDateRetrieval
}
