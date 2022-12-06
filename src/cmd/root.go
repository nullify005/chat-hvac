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
	"os"

	"github.com/spf13/cobra"
)

var (
	flagsSlackBotToken string
	flagsSlackAppToken string
	flagsSlackChannel  string
	flagsConfig        string
	flagsIntesis       string
	flagsDevice        string
	rootCmd            = &cobra.Command{
		Use:   "chat-hvac",
		Short: "A slack & service-intesis integration to control HVAC status",
	}
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagsSlackBotToken, "bottoken", "", "Slack BotToken xoxb-...")
	rootCmd.PersistentFlags().StringVar(&flagsSlackAppToken, "apptoken", "", "Slack AppToken xapp-...")
	rootCmd.PersistentFlags().StringVar(&flagsSlackChannel, "channel", "", "Slack Channel")
	rootCmd.PersistentFlags().StringVar(&flagsConfig, "config", "", "Path to Config file")
	rootCmd.PersistentFlags().StringVar(&flagsIntesis, "intesis", "", "URL for service-intesis endpoint")
	rootCmd.PersistentFlags().StringVar(&flagsDevice, "device", "", "Intesis device to conduct actions on")
	// if flagsConfig == "" {
	// 	rootCmd.MarkFlagRequired("bottoken")
	// 	rootCmd.MarkFlagRequired("apptoken")
	// 	rootCmd.MarkFlagRequired("channel")
	// } else {
	// 	rootCmd.MarkFlagRequired("config")
	// }
}