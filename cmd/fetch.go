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

type fetchCommand struct{}

// NewFetchCommand returns a Command that fetches an opaque object.
func NewFetchCommand() *fetchCommand {
	return &fetchCommand{}
}

// CLICommand implements Command.
func (c *fetchCommand) CLICommand() cli.Command {
	return cli.Command{
		Name:   "fetch",
		Usage:  "fetch opaque object contents with auth macaroon",
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
		},
	}
}

// Do implements Command.
func (c *fetchCommand) Do(ctx Context) error {
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

	var authBuf bytes.Buffer
	ms, err := unmarshalAuth(input)
	if err != nil {
		return err
	}
	ms, env, err := dischargeAuth(ctx, ms)
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

	req, err := http.NewRequest("POST", urlStr+"/"+id, bytes.NewBuffer(authBuf.Bytes()))
	if err != nil {
		return fmt.Errorf("failed to create request %q: %v", urlStr, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error requesting %q: %v", urlStr, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		var contents io.Reader
		if env != nil {
			contents, err = env.decrypt(resp.Body)
			if err != nil {
				return fmt.Errorf("error decrypting contents: %v", err)
			}
		} else {
			contents = resp.Body
		}
		_, err = io.Copy(output, contents)
		return err
	}
	return errHTTPResponse(resp)
}
