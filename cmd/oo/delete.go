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
	"gopkg.in/macaroon.v1"
)

func doDelete(c *cli.Context) {
	run(c, func(c *cli.Context) error {
		var (
			input io.ReadCloser
			err   error
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
		id, err := objectID(ms)
		if err != nil {
			return fmt.Errorf("cannot determine object ID: %v", err)
		}

		req, err := http.NewRequest("DELETE", urlStr+"/"+id, bytes.NewBuffer(mjson.Bytes()))
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
	})
}
