/*
Copyright © 2022 Lee Webb <nullify005@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/nullify005/chat-hvac/pkg/adapter"
	"github.com/nullify005/chat-hvac/pkg/adapter/console"
	"github.com/nullify005/chat-hvac/pkg/adapter/slack"
	"github.com/nullify005/chat-hvac/pkg/config"
	"github.com/nullify005/chat-hvac/pkg/health"
	"github.com/nullify005/chat-hvac/pkg/hvac"
	"github.com/nullify005/chat-hvac/pkg/receiver"
	"github.com/spf13/cobra"
)

var (
	flagsConfig  string
	flagsAdapter string
	rootCmd      = &cobra.Command{
		Use:   "chat-hvac",
		Short: "A slack & service-intesis integration to control HVAC status",
		Run: func(cmd *cobra.Command, args []string) {
			var adapter adapter.Adapter
			logger := log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile)
			logger.Print("logger is alive")
			c, err := config.New(flagsConfig)
			if err != nil {
				logger.Fatalf("unable to read config: %s cause: %v", flagsConfig, err)
			}
			if flagsAdapter == "slack" {
				adapter = slack.New(c.BotToken, c.AppToken, slack.WithLogger(logger))
			} else {
				adapter = console.New(console.WithLogger(logger))
			}
			h := hvac.New(hvac.WithApi(c.Intesis), hvac.WithDevice(c.Device), hvac.WithLogger(logger))
			r := receiver.New(adapter, receiver.WithLogger(logger), receiver.WithHvac(h))
			health := health.New(health.WithLogger(logger))
			health.Run()
			r.Receive()
		},
	}
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagsConfig, "config", "/.secrets/config.yaml", "Path to Config file")
	rootCmd.Flags().StringVarP(&flagsAdapter, "adapter", "a", "console", "The name of the IM adapter (slack|console) defaults to console")
}
