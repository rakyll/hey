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
	"os"
	"os/signal"
	"time"

	"github.com/angkeith/hey/internal/config"
	"github.com/angkeith/hey/internal/parser"
	"github.com/spf13/cobra"
)

var (
	conf    config.Config
	rootCmd = NewRootCmd()
)

func init() {
	InitFlags(rootCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "hey [flags] <url>",
		Version: config.Version,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf.Url = args[0]

			w, err := parser.NewWork(&conf)

			if err != nil {
				return err
			}

			w.Init()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				<-c
				w.Stop()
			}()
			if conf.Dur > 0 {
				go func() {
					time.Sleep(conf.Dur)
					w.Stop()
				}()
			}
			w.Run()
			return nil
		},
	}
}

func InitFlags(rootCmd *cobra.Command) {
	rootCmd.Flags().SortFlags = false
	// Disable help flag
	rootCmd.Flags().Bool("help", false, "")
	rootCmd.Flags().MarkHidden("help")

	// Define flags here.
	rootCmd.Flags().IntVarP(&conf.N, "number-of-requests", "n", config.NumberOfRequests, "Number of requests to run.")
	rootCmd.Flags().IntVarP(&conf.C, "concurrency", "c", config.Concurrency, `Number of workers to run concurrently. Total number of requests cannot
be smaller than the concurrency level.`)
	rootCmd.Flags().Float64VarP(&conf.Q, "rate-limit", "q", config.RateLimit, "Rate limit, in queries per second (QPS) per worker. Default is no rate limit.")
	rootCmd.Flags().DurationVarP(&conf.Dur, "duration", "z", config.Duration, `Duration of application to send requests. When duration is reached,
application stops and exits. If duration is specified, n is ignored.
Examples: -z 10s -z 3m.`)
	rootCmd.Flags().IntVar(&conf.Cpus, "cpus", config.Cpus, "Number of cpu cores to use.")
	rootCmd.Flags().StringVarP(&conf.Output, "output", "o", config.Output, `Output type. If none provided, a summary is printed.
"csv" is the only supported alternative. Dumps the response
metrics in comma-separated values format.
`)

	rootCmd.Flags().BoolVar(&conf.DisableCompression, "disable-compression", config.DisableCompression, "Disable compression.")
	rootCmd.Flags().BoolVar(&conf.DisableKeepAlives, "disable-keepalive", config.DisableKeepAlives, "Disable keep-alive, prevents re-use of TCP connections between different HTTP requests.")
	rootCmd.Flags().BoolVar(&conf.DisableRedirects, "disable-redirects", config.DisableRedirects, "Disable following of HTTP redirects.\n")

	rootCmd.Flags().StringVarP(&conf.M, "request", "X", config.Request, "HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.")
	rootCmd.Flags().StringArrayVarP(&conf.HeaderSlice, "header", "H", config.HeaderSlice, "Pass custom header to server, overriding any internal header.")
	rootCmd.Flags().IntVarP(&conf.T, "connect-timeout", "t", config.ConnectTimeout, "Maximum time in seconds allowed for a request to take.")
	// TODO: make this StringArray just like in curl
	rootCmd.Flags().StringVarP(&conf.Data, "data", "d", config.Data, "Sends the specified data in a POST requst to the HTTP server. If you start the data with \nthe letter @, the rest should be a file name to read the data from.")
	rootCmd.Flags().StringVarP(&conf.AuthHeader, "user", "u", config.User, "Server user and password")
	rootCmd.Flags().StringVarP(&conf.UserAgent, "user-agent", "A", config.UserAgent, "Send User-Agent Header to server.")
	rootCmd.Flags().StringVarP(&conf.ProxyAddr, "proxy", "x", config.Proxy, "HTTP Proxy address as host:port.")
	rootCmd.Flags().BoolVar(&conf.H2, "http2", config.Http2, "Use HTTP 2.")
	rootCmd.Flags().BoolVarP(&conf.Debug, "verbose", "v", config.Verbose, "Dumps request and response.")
}
