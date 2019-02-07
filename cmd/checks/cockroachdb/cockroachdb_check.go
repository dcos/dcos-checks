package cockroachdb

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dcos/dcos-checks/common"
	"github.com/dcos/dcos-checks/constants"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var validArgs = map[string]struct{}{
	"ranges": struct{}{},
}

// cockroachdbCmd represents the cluster command
var cockroachdbCmd = &cobra.Command{
	Use:   "cockroachdb",
	Short: "CockroachDB health checks",
	Long: `List of checks to verify CockroachDB is healthy/up
Usage:
cockroachdb ranges
`,
	Run: func(cmd *cobra.Command, args []string) {
		common.RunCheck(context.TODO(),
			newRangesCheck("DC/OS CockroachDB checks", args))
	},
}

// Register adds this command to the root command
func Register(root *cobra.Command) {
	root.AddCommand(cockroachdbCmd)
}

// newClusterLeaderCheck returns an intialized instance of *clusterLeaderCheck
func newRangesCheck(name string, args []string) *rangesCheck {
	return &rangesCheck{
		Name: name,
		Args: args,
	}
}

// rangesCheck verifies all CockroachDB ranges are fully
// replicated.
type rangesCheck struct {
	Name string
	Args []string
}

// ID returns a unique check identifier.
func (rc *rangesCheck) ID() string {
	return rc.Name
}

// Run the cluster check
func (rc *rangesCheck) Run(ctx context.Context, cfg *common.CLIConfigFlags) (string, int, error) {
	var args = rc.Args

	var keys []string
	for key := range validArgs {
		keys = append(keys, key)
	}

	if len(args) == 0 {
		return "", constants.StatusFailure, fmt.Errorf("No args provided, valid args %v", keys)
	}

	for _, arg := range args {
		switch arg {
		case "ranges":
			return rc.Ranges("127.0.0.1:8090", cfg)
		default:
		}
	}
	return "", constants.StatusFailure, fmt.Errorf("Option not supported, valid args %v", keys)
}

// Ranges checks that there are no unavailable or underreplicated ranges. If
// TLS is enforced the check does not perform TLS peer verification on the
// certificate presented by the remote endpoint.
func (rc *rangesCheck) Ranges(addr string, cfg *common.CLIConfigFlags) (string, int, error) {
	scheme := "http"
	var tlsConfig *tls.Config
	if cfg.ForceTLS {
		scheme = "https"
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	tr := &http.Transport{
		IdleConnTimeout: 10 * time.Second,
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{Transport: tr}
	u := url.URL{
		Scheme: scheme,
		Host:   addr,
		Path:   "/_status/nodes",
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", constants.StatusFailure, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", constants.StatusFailure, err
	}
	defer resp.Body.Close()
	var nodestats struct {
		Nodes []struct {
			StoreStatuses []struct {
				Metrics struct {
					RangesUnderreplicated int `json:"ranges.underreplicated"`
					RangesUnavailable     int `json:"ranges.unavailable"`
				} `json:"metrics"`
			} `json:"storeStatuses"`
			Args []string `json:"args"`
		} `json:"nodes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nodestats); err != nil {
		return "", constants.StatusFailure, err
	}
	for _, node := range nodestats.Nodes {
		underreplicated := 0
		unavailable := 0
		for _, storestat := range node.StoreStatuses {
			underreplicated += storestat.Metrics.RangesUnderreplicated
			unavailable += storestat.Metrics.RangesUnavailable
		}
		if unavailable > 0 {
			return "", constants.StatusFailure, errors.Errorf("CockroachDB has unavailable ranges")
		}
		if underreplicated > 0 {
			return "", constants.StatusFailure, errors.Errorf("CockroachDB has underreplicated ranges")
		}
	}
	return "", constants.StatusOK, nil
}
