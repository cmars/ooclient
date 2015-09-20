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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
)

var newCommand = cli.Command{
	Name:   "new",
	Usage:  "create a new opaque object with given input, output auth macaroon",
	Action: doNew,
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
			Name: "content-type, t",
		},
	},
}

func doNew(c *cli.Context) {
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

		req, err := http.NewRequest("POST", urlStr, input)
		if err != nil {
			return fmt.Errorf("failed to create request %q: %v", urlStr, err)
		}

		contentType := c.String("content-type")
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
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

func errHTTPResponse(resp *http.Response) error {
	var body bytes.Buffer
	_, err := io.Copy(&body, resp.Body)
	if err != nil {
		log.Println("error reading response: %v", err)
	}
	return fmt.Errorf("%s: %s", resp.Status, body.String())
}
