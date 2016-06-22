package list

import (
	//	"encoding/json"
	"fmt"
	//	"io/ioutil"
	//	"net/http"
	"net/url"
	//"os"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	//	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

type listOpts struct {
	registry string
	username string
	password string
}

func (opts *listOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

	baseURL, err := url.Parse(opts.registry)
	if err != nil {
		return err
	}

	// See if we can find some way to authenticate
	var auth registry.Authenticator = registry.NilAuth{}

	if opts.password != "" || opts.username != "" {
		auth = registry.BasicAuth{opts.username, opts.password}
	}

	client := registry.NewClient()
	client.BaseURL = baseURL

	tags, err := client.Repository.ListTags(image, auth)
	if err != nil {
		return err
	}
	for tag, _ := range tags {
		fmt.Println(tag)
	}
	return nil
}

func AddSubcommandTo(cmd *cobra.Command) {
	opts := &listOpts{}
	subcmd := &cobra.Command{
		Use:   "list <image>",
		Short: "list images",
		RunE:  opts.run,
	}
	subcmd.Flags().StringVar(&opts.registry, "image-registry", "https://index.docker.io/v1/", "image registry to query")
	subcmd.Flags().StringVar(&opts.username, "username", "", "username to use in basic authentication")
	subcmd.Flags().StringVar(&opts.password, "password", "", "password to use in basic authentication")

	cmd.AddCommand(subcmd)
}

// --- using the Docker config file to get auth

// const DockerConfigFile = "~/.docker/config.json"

// type token struct {
// 	Auth  string `json:"auth"`
// 	Email string `json:"email"`
// }

// type imageTag struct {
// 	Layer string `json:"layer"`
// 	Name  string `json:"name"`
// }

// type dockerConfig struct {
// 	Auths map[string]token `json:"auths"`
// }

// func findConfiguredAuth(config *dockerConfig, baseURL *url.URL) *token {
// 	for _, possibleKey := range []string{baseURL.String(), baseURL.Host} {
// 		if authEntry, hasConfig := config.Auths[possibleKey]; hasConfig {
// 			return &authEntry
// 		}
// 	}
// 	return nil
// }

// func loadDockerConfig() (*dockerConfig, error) {
// 	path, err := homedir.Expand(DockerConfigFile)
// 	if err != nil {
// 		panic(err)
// 	}
// 	bytes, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var config dockerConfig
// 	if err = json.Unmarshal(bytes, &config); err != nil {
// 		return nil, err
// 	}
// 	return &config, nil
// }
