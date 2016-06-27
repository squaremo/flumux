package list

import (
	//	"encoding/json"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	//	"io/ioutil"
	//	"net/http"
	"net/url"
	//"os"
	"sort"
	"text/tabwriter"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

const DockerConfigFile = "~/.docker/config.json"
const DockerHub = "https://index.docker.io/v1/"

type listOpts struct {
	registry        string
	gitrepo         string
	useDockerConfig bool
	username        string
	password        string
}

type resultEntry struct {
	tag      string
	msg      string
	commitID *git.Oid
}

type result struct {
	repo    *git.Repository
	entries []*resultEntry
}

func (result *result) Less(i, j int) bool {
	if result.entries[i].commitID == nil {
		return false
	} else if result.entries[j].commitID == nil {
		return true
	}
	// Define: A < B iff A is a descendant of B, i.e., comes after it in git history
	res, err := result.repo.DescendantOf(result.entries[i].commitID, result.entries[j].commitID)
	// assume an error indicates no relative ordering
	return (err == nil) && res
}

func (result *result) Swap(i, j int) {
	t := result.entries[i]
	result.entries[i] = result.entries[j]
	result.entries[j] = t
}

func (result *result) Len() int {
	return len(result.entries)
}

func (opts *listOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

	var (
		repo *git.Repository
		err  error
	)
	if opts.gitrepo != "" {
		repo, err = git.OpenRepository(opts.gitrepo)
		if err != nil {
			return err
		}
	}

	baseURL, err := url.Parse(opts.registry)
	if err != nil {
		return err
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
			return err
		}
	}

	tags, err := client.Repository.ListTags(image, auth)
	if err != nil {
		return err
	}

	if repo != nil {
		result := &result{repo, make([]*resultEntry, len(tags))}
		i := 0
		for tag, _ := range tags {
			entry := &resultEntry{tag: tag}
			result.entries[i] = entry
			i++

			additional := ""
			// hard-code tag format for now
			if strings.HasSuffix(tag, "-WIP") {
				additional = " (uncommitted changes)"
				tag = tag[:len(tag)-4]
			}
			commit, otherwise := commitFromTag(repo, tag)
			if otherwise != "" {
				entry.msg = otherwise
			} else {
				entry.commitID = commit.Id()
				entry.msg = strings.Split(commit.Message(), "\n")[0]
			}
			entry.msg = entry.msg + additional
		}
		sort.Sort(result)

		out := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
		for _, entry := range result.entries {
			fmt.Fprint(out, entry.tag)
			fmt.Fprint(out, "\t")
			fmt.Fprint(out, entry.msg)
			fmt.Fprint(out, "\n")
		}
		out.Flush()
	} else {
		for tag, _ := range tags {
			fmt.Println(tag)
		}
	}

	return nil
}

func commitFromTag(repo *git.Repository, tag string) (*git.Commit, string) {
	bits := strings.Split(tag, "-")
	if len(bits) != 2 {
		return nil, "tag does not correspond to a commit"
	}
	commitRev, err := repo.RevparseSingle(bits[1])
	if err != nil {
		return nil, err.Error()
	}
	commit, err := commitRev.AsCommit()
	if err != nil {
		return nil, err.Error()
	}
	return commit, ""
}

func AddSubcommandTo(cmd *cobra.Command) {
	opts := &listOpts{}
	subcmd := &cobra.Command{
		Use:   "list <image>",
		Short: "list images",
		RunE:  opts.run,
	}
	subcmd.Flags().StringVar(&opts.registry, "image-registry", DockerHub, "image registry to query")
	subcmd.Flags().StringVar(&opts.gitrepo, "repository", "", "path to git repository to cross-reference images to")
	subcmd.Flags().StringVar(&opts.username, "username", "", "username to use in basic authentication")
	subcmd.Flags().StringVar(&opts.password, "password", "", "password to use in basic authentication")
	subcmd.Flags().BoolVar(&opts.useDockerConfig, "use-docker-config", true, fmt.Sprintf("look up the authentication token in %s", DockerConfigFile))

	cmd.AddCommand(subcmd)
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
