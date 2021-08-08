// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/rakyll/hey/internal/config"
	"github.com/rakyll/hey/internal/parser"
	"github.com/rakyll/hey/requester"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var (
	work *requester.Work
	err  error
	out  string
)

func TestNoArgs(t *testing.T) {
	out, err = executeCommand("")
	require.NotEqual(t, err, nil, "expects error message")
	// expect print usage called once
}

func TestVerboseConf(t *testing.T) {
	_, err = executeCommand("-v https://example.com")
	expectedConf := config.NewConfigV("https://example.com")

	expectedConf.Debug = true
	expectedConf.C = 1
	expectedConf.N = 1
	fmt.Println(conf)
	fmt.Println(expectedConf)

	require.Equal(t, conf, expectedConf, "expects to inplictly set C and N to 1")
}

func TestNumberOfRequestsConf(t *testing.T) {
	_, err = executeCommand("-n 5 https://example.com")
	expectedConf := config.NewConfigV("https://example.com")
	expectedConf.N = 5
	expectedConf.C = expectedConf.N

	require.Equal(t, conf, expectedConf, "expects to inplictly set C to N when C > N")

	_, err = executeCommand("-n 1 -c 100 https://example.com")
	expectedConf = config.NewConfigV("https://example.com")
	expectedConf.N = 1
	expectedConf.C = expectedConf.N

	require.Equal(t, conf, expectedConf, "expects to inplictly set C to N when C > N")
}

// #################################################################################
// # helper
// #################################################################################
func executeCommand(s string) (output string, err error) {
	rootCmd := newCmdTestNewWork()
	_, output, err = executeCommandC(rootCmd, s)

	return output, err
}

func newCmdTestNewWork() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "hey [flags] <url>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf.Url = args[0]
			work, err = parser.NewWork(&conf)
			return err
		},
	}
	InitFlags(cmd)
	return cmd
}

func executeCommandC(root *cobra.Command, s string) (c *cobra.Command, output string, err error) {
	args := []string{}
	if s != "" {
		args = strings.Split(s, " ")
	}
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}
