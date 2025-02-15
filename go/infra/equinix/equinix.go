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

package equinix

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitessio/arewefastyet/go/infra"
	"github.com/vitessio/arewefastyet/go/infra/ansible"
	"log"
)

const (
	Name = "equinix"

	flagToken        = "equinix-token"
	flagProjectID    = "equinix-project-id"
	flagInstanceType = "equinix-instance-type"
)

type Equinix struct {
	Token     string
	ProjectID string
	InstanceType string
	tf        *tfexec.Terraform
	InfraCfg  *infra.Config
}

func (e *Equinix) AddToCommand(cmd *cobra.Command) {
	cmd.Flags().StringVar(&e.Token, flagToken, "", "Auth Token for Equinix Metal")
	cmd.Flags().StringVar(&e.ProjectID, flagProjectID, "", "Project ID to use for Equinix Metal")
	cmd.Flags().StringVar(&e.InstanceType, flagInstanceType, "", "Instance type to use for the creation of a new node")

	viper.BindPFlag(flagToken, cmd.Flags().Lookup(flagToken))
	viper.BindPFlag(flagProjectID, cmd.Flags().Lookup(flagProjectID))
	viper.BindPFlag(flagInstanceType, cmd.Flags().Lookup(flagInstanceType))
}

func (e Equinix) TerraformVarArray() (vars []*tfexec.VarOption) {
	vars = append(vars, tfexec.Var("auth_token=" + e.Token), tfexec.Var("project_id=" + e.ProjectID))
	if e.InstanceType != "" {
		vars = append(vars, tfexec.Var("instance_type="+e.InstanceType))
	}
	return vars
}

func (e Equinix) Create(wantOutputs ...string) (output map[string]string, err error) {
	if e.tf == nil {
		return nil, fmt.Errorf("%s: equinix terraform not prepared", infra.ErrorInvalidConfiguration)
	}

	planOpts := &[]tfexec.PlanOption{}
	if err = infra.PopulateTfOption(e.TerraformVarArray(), planOpts); err != nil {
		return nil, fmt.Errorf("%s: %s", infra.ErrorProvision, err.Error())
	}

	changed, err := e.tf.Plan(context.Background(), *planOpts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", infra.ErrorProvision, err.Error())
	} else if changed {
		log.Println("Applying tf plan.")
		applyOpts := &[]tfexec.ApplyOption{}
		if err = infra.PopulateTfOption(e.TerraformVarArray(), applyOpts); err != nil {
			return nil, fmt.Errorf("%s: %s", infra.ErrorProvision, err.Error())
		}
		err = e.tf.Apply(context.Background(), *applyOpts...)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", infra.ErrorProvision, err.Error())
		}
	} else {
		fmt.Println("plan did not change, no provision needed")
	}

	if len(wantOutputs) > 0 {
		tfOutput, err := e.tf.Output(context.Background())
		if err != nil {
			// todo: create error type
			return nil, err
		}
		output = make(map[string]string, len(wantOutputs))
		for _, wantOutput := range wantOutputs {
			output[wantOutput] = string(tfOutput[wantOutput].Value)
		}
	}

	return output, nil
}

func (e Equinix) ValidConfig() error {
	if e.Token == "" {
		return fmt.Errorf("%s: missing token", infra.ErrorInvalidConfiguration)
	} else if e.ProjectID == "" {
		return fmt.Errorf("%s: missing project id", infra.ErrorInvalidConfiguration)
	} else if err := e.InfraCfg.Valid(); err != nil {
		return err
	}
	return nil
}

func (e *Equinix) Prepare() error {
	if err := e.ValidConfig(); err != nil {
		return err
	}
	if err := e.InfraCfg.Prepare(); err != nil {
		return err
	}
	workingDir := e.InfraCfg.Path
	tf, err := tfexec.NewTerraform(workingDir, e.InfraCfg.PathExecTF)
	if err != nil {
		return err
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true), tfexec.LockTimeout("60s"))
	if err != nil {
		return err
	}

	err = tf.SetLogPath("trace.log")
	if err != nil {
		return err
	}

	e.tf = tf
	return nil
}

func (e *Equinix) Run(ansibleConfig *ansible.Config) error {
	return ansible.Run(ansibleConfig)
}

func (e *Equinix) SetConfig(config *infra.Config) {
	e.InfraCfg = config
}
