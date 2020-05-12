package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomocy/go-cookbook/oauth/authz/infra/memory"
	"github.com/tomocy/go-cookbook/oauth/authz/presentation/http"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	var conf config
	if err := conf.parse(args); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	var (
		clientRepo = memory.NewClientRepo()
		userRepo   = memory.NewUserRepo()
		sessRepo   = memory.NewSessionRepo()
	)

	serv, err := http.NewServer(w, conf.addr, clientRepo, userRepo, sessRepo)
	if err != nil {
		return fmt.Errorf("failed to craete server: %w", err)
	}
	if err := serv.Run(); err != nil {
		return fmt.Errorf("failed to run server: %s", err)
	}

	return nil
}

type config struct {
	addr string
}

func (c *config) parse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&c.addr, "addr", ":80", "the address to listen and serve")

	return flags.Parse(args[1:])
}