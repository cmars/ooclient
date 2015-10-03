/*
 * Copyright 2015 Casey Marshall
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
 */

package cmd

import (
	"fmt"

	"gopkg.in/basen.v1"

	"github.com/codegangsta/cli"
)

type keyCommand struct{}

// NewKeyCommand returns a Command that displays the client's public key.
func NewKeyCommand() *keyCommand {
	return &keyCommand{}
}

// CLICommand implements Command.
func (c *keyCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:   "key",
		Usage:  "display public key",
		Action: Action(c),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "home",
				EnvVar: "OO_HOME",
				Value:  defaultHome,
			},
		},
	}
}

// Do implements Command.
func (c *keyCommand) Do(ctx Context) error {
	mgr := keyManager{ctx}
	kp, err := mgr.keyPair()
	if err != nil {
		return fmt.Errorf("failed to load key: %v", err)
	}
	_, err = fmt.Println(basen.Base58.EncodeToString(kp.Public.Key[:]))
	return err
}
