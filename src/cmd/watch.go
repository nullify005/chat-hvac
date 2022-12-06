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
	"github.com/nullify005/chat-hvac/pkg/cli"
	"github.com/nullify005/chat-hvac/pkg/config"
	"github.com/spf13/cobra"
)

var (
	watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "watch",
		Run: func(cmd *cobra.Command, args []string) {
			c := &config.Config{}
			if flagsConfig != "" {
				c = cli.Config(flagsConfig)
			} else {
				c = &config.Config{
					AppToken: flagsSlackAppToken,
					BotToken: flagsSlackBotToken,
					Channel:  flagsSlackChannel,
					Intesis:  flagsIntesis,
					Device:   flagsDevice,
				}
			}
			cli.Watch(c.BotToken, c.AppToken, c.Channel, c.Intesis, c.Device)
		},
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)
}
