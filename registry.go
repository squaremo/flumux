package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

const DockerConfigFile = "~/.docker/config.json"
const DockerHub = "https://index.docker.io/v1/"

type registryOpts struct {
	registry        string
	useDockerConfig bool
	username        string
	password        string
}

type registryClient struct {
	*registry.Client
	auth registry.Authenticator
}

func (opts *listOpts) addRegistryFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opts.registry, "image-registry", DockerHub, "image registry to query")
	cmd.Flags().StringVar(&opts.username, "username", "", "username to use in basic authentication")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to use in basic authentication")
	cmd.Flags().BoolVar(&opts.useDockerConfig, "use-docker-config", true, fmt.Sprintf("look up the authentication token in %s", DockerConfigFile))
}

func (opts *registryOpts) newRegistryClient(image string) (*registryClient, error) {
	baseURL, err := url.Parse(opts.registry)
	if err != nil {
		return nil, err
	}
	// See if we can find some way to authenticate
	var auth registry.Authenticator = registry.NilAuth{}

	if opts.password != "" || opts.username != "" {
		auth = registry.BasicAuth{opts.username, opts.password}
	} else if opts.useDockerConfig {
		auth = findConfiguredAuth(baseURL)
	}

	client := registry.NewClient()
	client.BaseURL = baseURL

	// If it's Docker hub (and possibly others?), we have to go
	// through this extra step of getting a token
	if opts.registry == DockerHub {
		auth, err = client.Hub.GetReadTokenWithAuth(image, auth)
		if err != nil {
			return nil, err
		}
	}

	return &registryClient{Client: client, auth: auth}, nil
}

// --- using the Docker config file to get auth

type auth struct {
	Auth  string `json:"auth"`
	Email string `json:"email"`
}

type dockerConfig struct {
	Auths map[string]auth `json:"auths"`
}

func findConfiguredAuth(baseURL *url.URL) registry.Authenticator {
	var result registry.Authenticator = registry.NilAuth{}
	config, err := loadDockerConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: could not find and read Docker config file")
		return result
	}
	for _, possibleKey := range []string{baseURL.String(), baseURL.Host} {
		if entry, hasConfig := config.Auths[possibleKey]; hasConfig {
			asString, err := base64.StdEncoding.DecodeString(entry.Auth)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not decode auth entry for %s\n", baseURL)
				return result
			}
			asSlice := strings.SplitN(string(asString), ":", 2)
			result = &registry.BasicAuth{
				Username: asSlice[0],
				Password: asSlice[1],
			}
		}
	}
	return result
}

func loadDockerConfig() (*dockerConfig, error) {
	path, err := homedir.Expand(DockerConfigFile)
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config dockerConfig
	if err = json.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}