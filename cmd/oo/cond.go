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

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"gopkg.in/macaroon.v1"
)

var condCommand = cli.Command{
	Name:   "cond",
	Usage:  "place conditional caveats on auth macaroon",
	Action: doCond,
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

func doCond(c *cli.Context) {
	run(c, func(c *cli.Context) error {
		var (
			input  io.ReadCloser
			output io.WriteCloser
			err    error
		)

		inputFile := c.String("input")
		if inputFile == "" {
			input = os.Stdin
		} else {
			input, err = os.Open(inputFile)
			if err != nil {
				return fmt.Errorf("cannot open %q for input: %v", inputFile, err)
			}
			defer input.Close()
		}

		outputFile := c.String("output")
		if outputFile == "" {
			output = os.Stdout
		} else {
			output, err = os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("cannot create %q for output: %v", outputFile, err)
			}
			defer output.Close()
		}

		urlStr := c.String("url")
		if urlStr == "" {
			cli.ShowAppHelp(c)
			return errors.New("--url or OOSTORE_URL is required")
		}

		var mjson bytes.Buffer
		_, err = io.Copy(&mjson, input)
		if err != nil {
			return fmt.Errorf("failed to read input: %v", err)
		}
		var ms macaroon.Slice
		err = json.Unmarshal(mjson.Bytes(), &ms)
		if err != nil {
			return fmt.Errorf("failed to decode auth: %v", err)
		}
		if len(ms) == 0 {
			return fmt.Errorf("missing auth")
		}
		if !c.Args().Present() {
			cli.ShowAppHelp(c)
			return fmt.Errorf("missing condition arguments")
		}
		err = ms[0].AddFirstPartyCaveat(strings.Join(c.Args(), " "))
		if err != nil {
			return fmt.Errorf("failed to add caveat: %v", err)
		}

		err = json.NewEncoder(output).Encode(ms)
		if err != nil {
			return fmt.Errorf("failed to encode auth: %v", err)
		}
		return nil
	})
}
