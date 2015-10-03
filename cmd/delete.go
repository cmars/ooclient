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

	"github.com/codegangsta/cli"
)

type deleteCommand struct{}

// NewDeleteCommand returns a Command that deletes an opaque object.
func NewDeleteCommand() *deleteCommand {
	return &deleteCommand{}
}

// CLICommand implements Command.
func (c *deleteCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"del", "rm"},
		Usage:   "delete opaque object with auth macaroon",
		Action:  Action(c),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "url",
				EnvVar: "OOSTORE_URL",
			},
			cli.StringFlag{
				Name:   "home",
				EnvVar: "OO_HOME",
				Value:  defaultHome,
			},
			cli.StringFlag{
				Name: "input, i",
			},
		},
	}
}

// Do implements Command.
func (c *deleteCommand) Do(ctx Context) error {
	var (
		input io.ReadCloser
		err   error
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

	urlStr := ctx.String("url")
	if urlStr == "" {
		ctx.ShowAppHelp()
		return errors.New("--url or OOSTORE_URL is required")
	}

	var authBuf bytes.Buffer
	ms, err := unmarshalAuth(input)
	if err != nil {
		return err
	}
	ms, _, err = dischargeAuth(ctx, ms)
	if err != nil {
		return err
	}
	id, err := objectID(ms)
	if err != nil {
		return err
	}
	err = json.NewEncoder(&authBuf).Encode(ms)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", urlStr+"/"+id, bytes.NewBuffer(authBuf.Bytes()))
	if err != nil {
		return fmt.Errorf("failed to create request %q: %v", urlStr, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting %q: %v", urlStr, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		_, err = io.Copy(os.Stderr, resp.Body)
		return err
	}
	return errHTTPResponse(resp)
}
