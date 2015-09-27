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
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/codegangsta/cli"
	"gopkg.in/macaroon-bakery.v1/bakery"
)

var defaultHome string

func init() {
	var userHomeDir string
	if u, err := user.Current(); err == nil {
		userHomeDir = u.HomeDir
	}
	if userHomeDir == "" {
		userHomeDir = os.Getenv("HOME")
	}
	if userHomeDir != "" {
		defaultHome = filepath.Join(userHomeDir, ".oo")
	}
}

type initCommand struct{}

// NewInitCommand returns a Command that initializes a new key pair.
func NewInitCommand() *initCommand {
	return &initCommand{}
}

// CLICommand implements Command.
func (c *initCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:   "init",
		Usage:  "create a new key pair",
		Action: Action(c),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "home",
				EnvVar: "OO_HOME",
				Value:  defaultHome,
			},
			cli.BoolFlag{
				Name:  "overwrite",
				Usage: "overwrite any existing key pair",
			},
		},
	}
}

type keyPair struct {
	*bakery.KeyPair
}

func (kp *keyPair) load(keyPath string) error {
	f, err := os.Open(keyPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&kp.KeyPair)
}

func (kp *keyPair) save(keyPath string) error {
	f, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(kp.KeyPair)
}

// Do implements Command.
func (c *initCommand) Do(ctx Context) error {
	homeDir := filepath.FromSlash(ctx.String("home"))
	if homeDir == "" {
		return fmt.Errorf("could not determine OO_HOME, --home is required")
	}
	err := os.MkdirAll(homeDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create directory %q: %v", homeDir, err)
	}

	var exists bool
	keyPath := filepath.Join(homeDir, "key")
	if _, err := os.Stat(keyPath); err == nil {
		exists = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if exists && !ctx.Bool("overwrite") {
		return fmt.Errorf("key pair already exists, use --overwrite to replace it")
	}

	bakeryKeyPair, err := bakery.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to create new key pair: %v", err)
	}
	kp := keyPair{bakeryKeyPair}
	err = kp.save(keyPath)
	if err != nil {
		return fmt.Errorf("failed to save new key pair: %v", err)
	}
	_, err = fmt.Fprintf(ctx.Stdout(), "%s\n", kp.Public.String())
	return err
}
