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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/macaroon-bakery.v1/bakery"
	"gopkg.in/macaroon-bakery.v1/bakery/checkers"
	"gopkg.in/macaroon.v1"

	"github.com/codegangsta/cli"
)

type newCommand struct{}

// NewNewCommand returns a Command that creates a new opaque object.
func NewNewCommand() *newCommand {
	return &newCommand{}
}

// CLICommand implements Command.
func (c *newCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:   "new",
		Usage:  "create a new opaque object with given input, output auth macaroon",
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
				Name: "content-type",
			},
			cli.StringFlag{
				Name: "to, t",
			},
		},
	}
}

// Do implements Command.
func (c *newCommand) Do(ctx Context) error {
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

	env, input, err := encrypt(input)

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

	req, err := http.NewRequest("POST", urlStr, input)
	if err != nil {
		return fmt.Errorf("failed to create request %q: %v", urlStr, err)
	}

	contentType := ctx.String("content-type")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting %q: %v", urlStr, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		ms, err := unmarshalAuth(resp.Body)
		if err != nil {
			return fmt.Errorf("invalid auth response: %v", err)
		}
		err = newContext{ctx}.addThirdPartyCaveat(ms[0], env)
		if err != nil {
			return fmt.Errorf("failed to add third-party caveat: %v", err)
		}
		err = json.NewEncoder(output).Encode(ms)
		return err
	}
	return errHTTPResponse(resp)
}

type newContext struct {
	Context
}

func (ctx newContext) addThirdPartyCaveat(m *macaroon.Macaroon, env *envelope) error {
	condition, err := env.MarshalJSON()
	if err != nil {
		return err
	}
	mgr := keyManager{ctx.Context}
	kp, err := mgr.keyPair()
	if err != nil {
		return err
	}
	agent, err := bakery.NewService(bakery.NewServiceParams{
		Key:     kp.KeyPair,
		Locator: clientLocator{kp},
	})
	if err != nil {
		return err
	}
	return agent.AddCaveat(m, checkers.Caveat{Location: "client:encrypt", Condition: string(condition)})
}

type clientLocator struct {
	*keyPair
}

// PublicKeyForLocation implements bakery.PublicKeyLocator by providing the
// same initialized key every time.
// TODO: support multiple key identities, getting key from command line or something.
func (l clientLocator) PublicKeyForLocation(loc string) (*bakery.PublicKey, error) {
	return &l.KeyPair.Public, nil
}

func unmarshalAuth(r io.Reader) (macaroon.Slice, error) {
	var mjson bytes.Buffer
	_, err := io.Copy(&mjson, r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %v", err)
	}
	var ms macaroon.Slice
	err = json.Unmarshal(mjson.Bytes(), &ms)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth: %v", err)
	}
	return ms, nil
}
