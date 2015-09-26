package cmd

import (
	"log"

	"github.com/codegangsta/cli"
)

// Context defines the command-line flags and other parameters exposed from the
// command-line.
type Context interface {
	// Args returns a slice of string arguments after flags are parsed.
	Args() []string

	// ShowAppHelp prints command usage to the terminal.
	ShowAppHelp()

	// String returns the value specified for the given flag name, or empty
	// string if not set.
	String(flagName string) string
}

type Command interface {
	CLICommand() cli.Command
	Do(ctx Context) error
}

type context struct {
	ctx *cli.Context
}

func (ctx *context) Args() []string {
	return []string(ctx.ctx.Args())
}

func (ctx *context) String(flagName string) string {
	return ctx.ctx.String(flagName)
}

func (ctx *context) ShowAppHelp() {
	cli.ShowAppHelp(ctx.ctx)
}

func Action(command Command) func(*cli.Context) {
	return func(ctx *cli.Context) {
		err := command.Do(&context{ctx})
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}
