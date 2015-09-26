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

package cmd_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/cmars/oostore"
	gc "gopkg.in/check.v1"
	"gopkg.in/tomb.v2"

	"github.com/cmars/ooclient/cmd"
)

func Test(t *testing.T) { gc.TestingT(t) }

type cmdSuite struct {
	server *httptest.Server
}

var _ = gc.Suite(&cmdSuite{})

func (s *cmdSuite) SetUpTest(c *gc.C) {
	store := oostore.NewMemStorage()
	service, err := oostore.NewService(oostore.ServiceConfig{
		ObjectStore: store,
	})
	c.Assert(err, gc.IsNil)
	s.server = httptest.NewServer(service)
}

func (s *cmdSuite) TearDownTest(c *gc.C) {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *cmdSuite) TestNewFetch(c *gc.C) {
	in := bytes.NewBufferString("hello world")
	var out bytes.Buffer
	fetchIn, newOut := io.Pipe()

	newCtx := &StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: in, stdout: newOut,
	}
	newCmd := cmd.NewNewCommand()
	fetchCtx := &StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: fetchIn, stdout: &out,
	}
	fetchCmd := cmd.NewFetchCommand()

	var t tomb.Tomb
	t.Go(func() error {
		return newCmd.Do(newCtx)
	})
	t.Go(func() error {
		return fetchCmd.Do(fetchCtx)
	})
	c.Assert(t.Wait(), gc.IsNil)
	c.Assert(out.String(), gc.Equals, "hello world")
}

func (s *cmdSuite) TestDelete(c *gc.C) {
	in := bytes.NewBufferString("hello world")
	var out bytes.Buffer

	// create
	c.Assert(cmd.NewNewCommand().Do(&StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: in, stdout: &out,
	}), gc.IsNil)
	// then delete
	c.Assert(cmd.NewDeleteCommand().Do(&StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: bytes.NewBuffer(out.Bytes()),
	}), gc.IsNil)
	// now it's gone
	c.Assert(cmd.NewFetchCommand().Do(&StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: bytes.NewBuffer(out.Bytes()),
	}), gc.ErrorMatches, `^404 Not Found.*`)
	c.Assert(cmd.NewDeleteCommand().Do(&StubContext{
		flags: map[string]interface{}{"url": s.server.URL},
		stdin: bytes.NewBuffer(out.Bytes()),
	}), gc.ErrorMatches, `^404 Not Found.*`)
}

// StubContext implements cmd.Context for stub testing purposes.
type StubContext struct {
	args   []string
	flags  map[string]interface{}
	stdin  io.Reader
	stdout io.Writer
}

func (c *StubContext) Args() []string {
	return c.args
}

func (c *StubContext) ShowAppHelp() {
	panic("help")
}

func (c *StubContext) String(flagName string) string {
	if c.flags == nil {
		return ""
	}
	val := c.flags[flagName]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (c *StubContext) Stdin() io.ReadCloser {
	if c.stdin == nil {
		return ioutil.NopCloser(bytes.NewBuffer(nil))
	}
	if rc, ok := c.stdin.(io.ReadCloser); ok {
		return rc
	}
	return ioutil.NopCloser(c.stdin)
}

type nopWriteCloser struct {
	io.Writer
}

func (c *nopWriteCloser) Close() error {
	return nil
}

func (c *StubContext) Stdout() io.WriteCloser {
	if c.stdout == nil {
		return &nopWriteCloser{ioutil.Discard}
	}
	if wc, ok := c.stdout.(io.WriteCloser); ok {
		return wc
	}
	return &nopWriteCloser{c.stdout}
}
