// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

package portavailability

import (
	"context"
	"fmt"
	"net"

	"github.com/dcos/dcos-checks/common"
	"github.com/dcos/dcos-checks/constants"
	"github.com/spf13/cobra"
)

// portAvailabilityCheck checks if ports are available
type portAvailabilityCheck struct {
	Name string
	Args []string
}

// portAvailabilityCmd represents the portavailability command
var portAvailabilityCmd = &cobra.Command{
	Use:   "port-availability",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("port-availability called")
	},
}

// Add adds this command to the root command
func Add(root *cobra.Command) {
	root.AddCommand(portAvailabilityCmd)
}

// newPortAvailabilityCheck returns an initialized instance of *portAvailabilityCheck.
func newPortAvailabilitysCheck(name string, args []string) common.DCOSChecker {
	return &portAvailabilityCheck{
		Name: name,
		Args: args,
	}
}

// ID returns a unique check identifier.
func (pa *portAvailabilityCheck) ID() string {
	return pa.Name
}

// portAvailable checks if a port is available
func (pa *portAvailabilityCheck) portAvailable(port string) (string, int, error) {
	ln, err := net.Listen("tcp", ":"+port)

	if err != nil {
		return "Cannot listen on port", constants.StatusFailure, err
	}

	err = ln.Close()
	if err != nil {
		return "Cannot close port connection", constants.StatusFailure, err
	}

	return "", constants.StatusOK, nil
}

// Run invokes a check and return error output, exit code and error.
func (pa *portAvailabilityCheck) Run(ctx context.Context, cfg *common.CLIConfigFlags) (string, int, error) {
	var args = pa.Args

	if len(args) == 0 {
		return "No port to check", constants.StatusFailure, nil
	}

	if len(args) > 1 {
		return "Only one port allowed", constants.StatusFailure, nil
	}

	port := args[0]
	return pa.portAvailable(port)
}
