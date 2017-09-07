// Copyright © 2017 Mesosphere
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dcos/dcos-go/dcos"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"reflect"
	"sort"
)

// NewVersionCheck returns an initialized instance of *VersionCheck.
func NewVersionCheck(name string) DCOSChecker {
	check := &VersionCheck{Name: name}
	check.ClusterLeader = dcos.DNSRecordLeader
	return check
}

// MasterListResponse response for leader.mesos/master.mesos
type MasterListResponse []struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
}

// AgentListResponse response for /slaves
type AgentListResponse struct {
	Slaves []struct {
		ID         string `json:"id"`
		Hostname   string `json:"hostname"`
		Port       int    `json:"port"`
		Attributes struct {
			PublicIP string `json:"public_ip"`
		} `json:"attributes"`
	} `json:"slaves"`
	RecoveredSlaves []interface{} `json:"recovered_slaves"`
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Check DC/OS version of the cluster",
	Long: `Check dc/os version on each node in the cluster.
At any point there shouldnt be more than 2 versions that exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunCheck(context.TODO(), NewVersionCheck("DC/OS version check"))
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

// VersionCheck struct
type VersionCheck struct {
	Name          string
	ClusterLeader string
}

// ID returns a unique check identifier.
func (vc *VersionCheck) ID() string {
	return vc.Name
}

// Run is running
func (vc *VersionCheck) Run(ctx context.Context, cfg *CLIConfigFlags) (string, int, error) {

	// List of masters
	var masterOpt URLFields
	masterOpt.host = "master.mesos"
	masterOpt.port = 80
	if cfg.ForceTLS {
		masterOpt.port = adminrouterMasterHTTPSPort
	}
	masterOpt.path = "/mesos_dns/v1/hosts/master.mesos"

	masterList, err := vc.ListOfMasters(cfg, masterOpt)
	if err != nil {
		return "", statusFailure, err
	}

	// List of agents
	var agentOpt URLFields
	agentOpt.host = "master.mesos"
	agentOpt.port = 80
	if cfg.ForceTLS {
		agentOpt.port = adminrouterMasterHTTPSPort
	}
	agentOpt.path = "/mesos/master/slaves"

	agentList, err := vc.ListOfAgents(cfg, agentOpt)
	if err != nil {
		return "", statusFailure, err
	}

	var versionURL URLFields
	versionURL.path = "/pkgpanda/active/"

	var lists [][]string
	m := make(map[string][]string)
	m["master"] = masterList
	m["agent"] = agentList

	for role, listOfHosts := range m {
		for _, host := range listOfHosts {
			versionURL.host = host
			if role == dcos.RoleMaster {
				versionURL.port = 80
				if cfg.ForceTLS {
					versionURL.port = adminrouterMasterHTTPSPort
				}
			}

			if role == dcos.RoleAgent {
				versionURL.port = adminrouterAgentHTTPPort
				if cfg.ForceTLS {
					versionURL.port = adminrouterAgentHTTPSPort
				}
			}

			packageList, err := vc.GetVersion(cfg, versionURL)
			if err != nil {
				return "", statusFailure, errors.Wrap(err, "Unable to get version")
			}
			sort.Strings(packageList)
			if len(lists) == 0 {
				// First time we got a package list
				lists = append(lists, packageList)
				continue
			}

			for _, list := range lists {
				sort.Strings(packageList)
				if !reflect.DeepEqual(list, packageList) {
					lists = append(lists, packageList)
				}
			}

			if len(lists) > 2 {
				err := fmt.Errorf("More than 2 dcos versions on the cluster")
				return "", statusFailure, err
			}
		}
	}

	return "", statusOK, nil
}

// ListOfMasters returns the current list of masters in the cluster
func (vc *VersionCheck) ListOfMasters(cfg *CLIConfigFlags, urlopt URLFields) ([]string, error) {
	var masterResponse MasterListResponse
	var masterIPs []string
	_, response, err := HTTPRequest(cfg, urlopt)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch list of masters")
	}

	if err := json.Unmarshal(response, &masterResponse); err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal response %s", string(response))
	}

	for _, addr := range masterResponse {
		masterIPs = append(masterIPs, addr.IP)
	}
	return masterIPs, nil
}

// ListOfAgents returns the current list of agents in the cluster
func (vc *VersionCheck) ListOfAgents(cfg *CLIConfigFlags, urlopt URLFields) ([]string, error) {
	var agentResponse AgentListResponse
	var agentIPs []string
	_, response, err := HTTPRequest(cfg, urlopt)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to fetch list of agents")
	}

	if err := json.Unmarshal(response, &agentResponse); err != nil {
		return nil, errors.Wrapf(err, "Unable to unmarshal response %s", string(response))
	}

	for _, hosts := range agentResponse.Slaves {
		agentIPs = append(agentIPs, hosts.Hostname)
	}
	return agentIPs, nil
}

// GetVersion returns the dc/os version of a node
func (vc *VersionCheck) GetVersion(cfg *CLIConfigFlags, urlopt URLFields) ([]string, error) {
	var verResponse []string
	_, response, err := HTTPRequest(cfg, urlopt)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get version")
	}

	if err := json.Unmarshal(response, &verResponse); err != nil {
		return nil, errors.Wrapf(err, "Unable to umarshal response, %s %v", string(response), urlopt)
	}

	return verResponse, nil
}
