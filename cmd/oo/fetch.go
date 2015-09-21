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
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"gopkg.in/macaroon-bakery.v1/bakery/checkers"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
	"gopkg.in/macaroon.v1"
)

var fetchCommand = cli.Command{
	Name:   "fetch",
	Usage:  "fetch opaque object contents with auth macaroon",
	Action: doFetch,
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

func doFetch(c *cli.Context) {
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

		auth, id, err := readAuth(input)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", urlStr+"/"+id, bytes.NewBuffer(auth))
		if err != nil {
			return fmt.Errorf("failed to create request %q: %v", urlStr, err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("error requesting %q: %v", urlStr, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			_, err = io.Copy(output, resp.Body)
			return err
		}
		return errHTTPResponse(resp)
	})
}

func readAuth(r io.Reader) ([]byte, string, error) {
	var fail string
	var mjson bytes.Buffer
	_, err := io.Copy(&mjson, r)
	if err != nil {
		return nil, fail, fmt.Errorf("failed to read input: %v", err)
	}
	var ms macaroon.Slice
	err = json.Unmarshal(mjson.Bytes(), &ms)
	if err != nil {
		return nil, fail, fmt.Errorf("failed to decode auth: %v", err)
	}
	id, err := objectID(ms)
	if err != nil {
		return nil, fail, fmt.Errorf("cannot determine object ID: %v", err)
	}

	if len(ms) == 1 {
		// TODO: this doesn't really address the case where the client obtains
		// some of the needed third-party discharges, but not all of them.
		cl := httpbakery.NewClient()
		ms, err = cl.DischargeAll(ms[0])
		if err != nil {
			return nil, fail, fmt.Errorf("failed to discharge third-party caveat: %v", err)
		}
		mjson.Reset()
		err = json.NewEncoder(&mjson).Encode(ms)
		if err != nil {
			return nil, fail, fmt.Errorf("failed to encode macaroon with discharges")
		}
	}

	return mjson.Bytes(), id, nil
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
