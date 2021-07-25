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
	"runtime"
	"time"

	"github.com/rakyll/hey/internal/parser"
	"github.com/spf13/cobra"
)

var conf parser.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:  "hey [flags] <url>",
	Args: cobra.ExactArgs(1),
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().SortFlags = false
	// Disable help flag
	rootCmd.Flags().Bool("help", false, "")
	rootCmd.Flags().MarkHidden("help")

	// Define flags here.
	rootCmd.Flags().IntVarP(&conf.N, "number-of-requests", "n", 200, "Number of requests to run.")
	rootCmd.Flags().IntVarP(&conf.C, "concurrency", "c", 50, `Number of workers to run concurrently. Total number of requests cannot
be smaller than the concurrency level.`)
	rootCmd.Flags().Float64VarP(&conf.Q, "rate-limit", "q", 0, "Rate limit, in queries per second (QPS) per worker. Default is no rate limit.")
	rootCmd.Flags().DurationVarP(&conf.Dur, "duration", "z", 0, `Duration of application to send requests. When duration is reached,
application stops and exits. If duration is specified, n is ignored.
Examples: -z 10s -z 3m.`)
	rootCmd.Flags().StringVarP(&conf.Output, "output", "o", "", `Output type. If none provided, a summary is printed.
"csv" is the only supported alternative. Dumps the response
metrics in comma-separated values format.
`)

	rootCmd.Flags().StringVarP(&conf.M, "request-method", "m", "GET", "HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.")
	rootCmd.Flags().StringArrayVarP(&conf.HeaderSlice, "header", "H", nil, `Custom HTTP header. You can specify as many as needed by repeating the flag.
For example, -H "Accept: text/html" -H "Content-Type: application/xml". `)
	rootCmd.Flags().IntVarP(&conf.T, "timeout", "t", 20, "Timeout for each request in seconds. Use 0 for infinite.")

	rootCmd.Flags().StringVarP(&conf.Accept, "accept", "A", "", "HTTP Accept header.")
	rootCmd.Flags().StringVarP(&conf.Body, "body", "d", "", "HTTP request body.")
	rootCmd.Flags().StringVarP(&conf.BodyFile, "body-file", "D", "", "HTTP request body from file. For example, /home/user/file.txt or ./file.txt.")
	rootCmd.Flags().StringVarP(&conf.ContentType, "content-type", "T", "text/html", "Content-type.")
	rootCmd.Flags().StringVarP(&conf.UserAgent, "user-agent", "U", "", `User-Agent, defaults to version "hey/0.0.1".`)
	rootCmd.Flags().StringVarP(&conf.AuthHeader, "auth-header", "a", "", "Basic authentication, username:password.")
	rootCmd.Flags().StringVarP(&conf.ProxyAddr, "proxy", "x", "", "HTTP Proxy address as host:port.")
	rootCmd.Flags().BoolVar(&conf.H2, "h2", false, "Enable HTTP/2.\n")

	rootCmd.Flags().StringVar(&conf.HostHeader, "host", "", "HTTP Host header.\n")

	rootCmd.Flags().BoolVar(&conf.DisableCompression, "disable-compression", false, "Disable compression.")
	rootCmd.Flags().BoolVar(&conf.DisableKeepAlives, "disable-keepalive", false, "Disable keep-alive, prevents re-use of TCP connections between different HTTP requests.")
	rootCmd.Flags().BoolVar(&conf.DisableRedirects, "disable-redirects", false, "Disable following of HTTP redirects.")
	rootCmd.Flags().IntVar(&conf.Cpus, "cpus", runtime.GOMAXPROCS(-1), "Number of cpu cores to use.")
}
