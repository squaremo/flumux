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

func imageName(image, tag string) string {
	return fmt.Sprintf("%s:%s", image, tag)
}

const (
	dockerConfigFile = "~/.docker/config.json"
	dockerHubHost    = "index.docker.io"
	dockerHubLibrary = "library"
)

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

func (opts *registryOpts) addRegistryFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opts.username, "username", "", "username to use in basic authentication")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to use in basic authentication")
	cmd.Flags().BoolVar(&opts.useDockerConfig, "use-docker-config", true, fmt.Sprintf("look up the authentication token in %s", dockerConfigFile))
}

func imageParts(repo string) (host, image, tag string, err error) {
	var org, imageAndTag string
	parts := strings.Split(repo, "/")
	switch len(parts) {
	case 1:
		host = dockerHubHost
		org = dockerHubLibrary
		imageAndTag = parts[0]
	case 2:
		host = dockerHubHost
		org = parts[0]
		imageAndTag = parts[1]
	case 3:
		host = parts[0]
		org = parts[1]
		imageAndTag = parts[2]
	default:
		return "", "", "", fmt.Errorf(`expected image name as either "host/org/image[:tag]", "org/image[:tag]", or "image:[tag]"`)
	}

	imageParts := strings.SplitN(imageAndTag, ":", 2)
	switch len(imageParts) {
	case 1:
		image = org + "/" + imageAndTag
	case 2:
		image = org + "/" + imageParts[0]
		tag = imageParts[1]
	}
	return
}

func (opts *registryOpts) newRegistryClient(repository string) (*registryClient, error) {
	host, image, _, err := imageParts(repository)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse image repository %s", repository)
	}

	baseURLStr := "https://" + host + "/v1/"
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("Somehow failed to parse URL I constructed myself: %s", baseURLStr)
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

	auth, err = client.Hub.GetReadTokenWithAuth(image, auth)
	if err != nil {
		return nil, err
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
	path, err := homedir.Expand(dockerConfigFile)
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
