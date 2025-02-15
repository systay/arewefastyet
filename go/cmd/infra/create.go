/*
 *
 * Copyright 2021 The Vitess Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package infra

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vitessio/arewefastyet/go/infra"
	"github.com/vitessio/arewefastyet/go/infra/equinix"
)

func create(cfg *infra.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a new instance",
	}

	cmd.AddCommand(createEquinix(cfg))
	return cmd
}

func createEquinix(cfg *infra.Config) *cobra.Command {
	eq := equinix.Equinix{InfraCfg: cfg}

	cmd := &cobra.Command{
		Use:     "equinix",
		Aliases: []string{"e"},
		Short:   "Create an Equinix Metal instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer cfg.Close()

			if err := eq.Prepare(); err != nil {
				return err
			}
			out, err := eq.Create("device_public_ip")
			fmt.Printf("Device IP: %s\n", out["device_public_ip"])
			return err
		},
	}
	eq.AddToCommand(cmd)
	return cmd
}
