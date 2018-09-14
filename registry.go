package main

import (
	"fmt"
	"io/ioutil"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/weaveworks/flux/image"
	"github.com/weaveworks/flux/registry"
	"github.com/weaveworks/flux/registry/middleware"
)

const DockerConfigFile = "~/.docker/config.json"

type registryOpts struct {
	dockerConfig string
	username     string
	password     string
}

func (opts *registryOpts) addRegistryFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&opts.username, "username", "", "username to use in basic authentication")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to use in basic authentication")
	cmd.Flags().StringVar(&opts.dockerConfig, "docker-config", DockerConfigFile, fmt.Sprintf("look up the authentication token in %s; supply an empty string to disable", DockerConfigFile))
}

func (opts *registryOpts) newRegistryClient(image image.Ref) (registry.Client, error) {
	clientFactory := &registry.RemoteClientFactory{
		Logger: nil,
		Limiters: &middleware.RateLimiters{
			RPS:   100,
			Burst: 10,
		},
	}

	auth, err := opts.findConfiguredAuth()
	if err != nil {
		return nil, err
	}
	return clientFactory.ClientFor(image.CanonicalName(), auth)
}

// --- using the Docker config file to get auth

func (opts *registryOpts) findConfiguredAuth() (registry.Credentials, error) {
	if opts.dockerConfig != "" {
		path, err := homedir.Expand(DockerConfigFile)
		if err != nil {
			return registry.Credentials{}, err
		}
		bytes, err := ioutil.ReadFile(path)
		if creds, err := registry.ParseCredentials(path, bytes); err != nil {
			return creds, nil
		}
	}
	return registry.NoCredentials(), nil
}

func parseImage(im string) (image.Ref, error) {
	return image.ParseRef(im)
}

// Make an image name from a base image reference or name, and a tag
func imageName(im image.Ref, tag string) string {
	return im.ToRef(tag).String()
}
