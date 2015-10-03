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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"gopkg.in/macaroon-bakery.v1/bakery"
	"gopkg.in/macaroon-bakery.v1/bakery/checkers"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"gopkg.in/macaroon.v1"
)

// Context defines the command-line flags and other parameters exposed from the
// command-line.
type Context interface {
	// Args returns a slice of string arguments after flags are parsed.
	Args() []string

	// Bool returns the boolean value specified for the given flag name.
	Bool(flagName string) bool

	// ShowAppHelp prints command usage to the terminal.
	ShowAppHelp()

	// String returns the value specified for the given flag name, or empty
	// string if not set.
	String(flagName string) string

	// Stdin returns the reader from standard input.
	Stdin() io.ReadCloser

	// Stdout returns the writer to standard output.
	Stdout() io.WriteCloser
}

// Command defines an ooclient subcommand.
type Command interface {
	// CLICommand returns an initialized cli.Command.
	CLICommand() cli.Command

	// Do implements the command action.
	Do(ctx Context) error
}

type context struct {
	ctx *cli.Context
}

// Args implements Context.
func (ctx *context) Args() []string {
	return []string(ctx.ctx.Args())
}

// Bool implements Context.
func (ctx *context) Bool(flagName string) bool {
	return ctx.ctx.Bool(flagName)
}

// ShowAppHelp implements Context.
func (ctx *context) ShowAppHelp() {
	cli.ShowAppHelp(ctx.ctx)
}

// String implements Context.
func (ctx *context) String(flagName string) string {
	return ctx.ctx.String(flagName)
}

// Stdin implements Context.
func (ctx *context) Stdin() io.ReadCloser {
	return os.Stdin
}

// Stdout implements Context.
func (ctx *context) Stdout() io.WriteCloser {
	return os.Stdout
}

func homeDir(ctx Context) (string, error) {
	home := filepath.FromSlash(ctx.String("home"))
	if home == "" {
		return "", fmt.Errorf("could not determine OO_HOME, --home is required")
	}
	return home, nil
}

// Action wraps a Command with a function that can be used with the cli
// package.
func Action(command Command) func(*cli.Context) {
	return func(ctx *cli.Context) {
		err := command.Do(&context{
			ctx: ctx,
		})
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}

func errHTTPResponse(resp *http.Response) error {
	var body bytes.Buffer
	_, err := io.Copy(&body, resp.Body)
	if err != nil {
		log.Println("error reading response: %v", err)
	}
	return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(body.String()))
}

type dischargeAcquirer struct {
	client *httpbakery.Client
	env    *envelope
}

// AcquireDischarge implements httpbakery.DischargeAcquirer.
func (da *dischargeAcquirer) AcquireDischarge(firstPartyLocation string, cav macaroon.Caveat) (*macaroon.Macaroon, error) {
	if cav.Location == "client:encrypt" {
		dm, _, err := bakery.Discharge(da.client.Key,
			bakery.ThirdPartyCheckerFunc(da.clientEncryptChecker), cav.Id)
		return dm, err
	}
	return da.client.AcquireDischarge(firstPartyLocation, cav)
}

func (da *dischargeAcquirer) clientEncryptChecker(caveatId, caveat string) ([]checkers.Caveat, error) {
	da.env = newEnvelope()
	err := da.env.UnmarshalJSON([]byte(caveat))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func dischargeAuth(ctx Context, ms macaroon.Slice) (macaroon.Slice, *envelope, error) {
	if len(ms) != 1 {
		return ms, nil, nil
	}
	mgr := keyManager{ctx}
	kp, err := mgr.keyPair()
	if err != nil {
		return nil, nil, err
	}

	cl := httpbakery.NewClient()
	da := &dischargeAcquirer{client: cl}
	cl.DischargeAcquirer = da
	cl.Key = kp.KeyPair
	ms, err = cl.DischargeAll(ms[0])
	if err != nil {
		return nil, nil, err
	}
	return ms, da.env, nil
}

func objectID(ms macaroon.Slice) (string, error) {
	var fail string
	var id string
	for _, m := range ms {
		for _, cav := range m.Caveats() {
			cond, arg, err := checkers.ParseCaveat(cav.Id)
			if err != nil {
				// strange, but offtopic
				continue
			}
			if cond == "object" {
				if id == "" {
					id = arg
				} else {
					return fail, fmt.Errorf("multiple conflicting caveats")
				}
			}
		}
	}
	if id == "" {
		return fail, errors.New("not found")
	}
	return id, nil
}
