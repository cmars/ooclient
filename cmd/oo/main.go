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
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func run(c *cli.Context, f func(*cli.Context) error) {
	err := f(c)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "oo"
	app.Usage = "oo [command] [args]"
	app.Commands = []cli.Command{{
		Name:   "new",
		Usage:  "oo new [-i|--input file] [-o|--output file] [-t|--content-type type]",
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
	}, {
		Name:   "fetch",
		Usage:  "oo fetch [-i|--input file] [-o|--output file]",
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
	}, {
		Name:   "cond",
		Usage:  "oo cond [-i|--input file] [-o|--output file] condition",
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
	}, {
		Name:    "delete",
		Aliases: []string{"del", "rm"},
		Usage:   "oo delete [-i|--input file]",
		Action:  doDelete,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "url",
				EnvVar: "OOSTORE_URL",
			},
			cli.StringFlag{
				Name: "input, i",
			},
		},
	}}
	app.Run(os.Args)
}
