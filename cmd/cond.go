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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"gopkg.in/macaroon-bakery.v1/bakery"
	"gopkg.in/macaroon-bakery.v1/bakery/checkers"
	"gopkg.in/macaroon.v1"
)

type condCommand struct{}

// NewCondCommand returns a Command that attenuates an opaque object auth with
// caveat conditions.
func NewCondCommand() *condCommand {
	return &condCommand{}
}

// CLICommand implements Command.
func (c *condCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:   "cond",
		Usage:  "place conditional caveats on auth macaroon",
		Action: Action(c),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "url",
				EnvVar: "OOSTORE_URL",
			},
			cli.StringFlag{
				Name: "input, i",
			},
			cli.StringFlag{
				Name: "output, o",
			},
			cli.StringFlag{
				Name:  "location, loc, l",
				Usage: "location of service for third-party caveat",
			},
			cli.StringFlag{
				Name:  "key, k",
				Usage: "base64-encoded public key of third-party service",
			},
		},
	}
}

// Do implements Command.
func (c *condCommand) Do(ctx Context) error {
	var (
		input  io.ReadCloser
		output io.WriteCloser
		err    error
	)

	inputFile := ctx.String("input")
	if inputFile == "" {
		input = ctx.Stdin()
	} else {
		input, err = os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("cannot open %q for input: %v", inputFile, err)
		}
	}
	defer input.Close()

	outputFile := ctx.String("output")
	if outputFile == "" {
		output = ctx.Stdout()
	} else {
		output, err = os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("cannot create %q for output: %v", outputFile, err)
		}
	}
	defer output.Close()

	urlStr := ctx.String("url")
	if urlStr == "" {
		ctx.ShowAppHelp()
		return errors.New("--url or OOSTORE_URL is required")
	}

	ms, err := unmarshalAuth(input)
	if err != nil {
		return fmt.Errorf("failed to unmarshal auth: %v", err)
	}
	if len(ms) == 0 {
		return fmt.Errorf("missing auth")
	}
	if len(ctx.Args()) == 0 {
		ctx.ShowAppHelp()
		return fmt.Errorf("missing condition arguments")
	}

	condition := strings.Join(ctx.Args(), " ")

	location := ctx.String("location")
	if location == "" {
		err = ms[0].AddFirstPartyCaveat(condition)
		if err != nil {
			return fmt.Errorf("failed to add caveat: %v", err)
		}
	} else {
		err = condContext{ctx}.addThirdPartyCaveat(ms[0], location, condition)
		if err != nil {
			return fmt.Errorf("failed to add caveat: %v", err)
		}
	}

	err = json.NewEncoder(output).Encode(ms)
	if err != nil {
		return fmt.Errorf("failed to encode auth: %v", err)
	}
	return nil
}

type condContext struct {
	Context
}

func (ctx condContext) addThirdPartyCaveat(m *macaroon.Macaroon, location, condition string) error {
	agent, err := bakery.NewService(bakery.NewServiceParams{
		// TODO: persistent key pair for client
		Locator: ctx,
	})
	if err != nil {
		return err
	}
	return agent.AddCaveat(m, checkers.Caveat{Location: location, Condition: condition})
}

// PublicKeyForLocation implements bakery.PublicKeyLocator by providing the key
// that was specified on the command line.
// TODO: PKIWTFBBQ.
// TODO: request keys on-demand if location is HTTPS.
func (ctx condContext) PublicKeyForLocation(loc string) (*bakery.PublicKey, error) {
	var key bakery.Key
	keyText := ctx.String("key")
	if keyText == "" {
		return nil, fmt.Errorf("--key is required for third-party caveat")
	}
	err := key.UnmarshalText([]byte(keyText))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal key %q: %v", keyText, err)
	}
	return &bakery.PublicKey{key}, nil
}
