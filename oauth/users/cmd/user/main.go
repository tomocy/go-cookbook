package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomocy/go-cookbook/oauth/users/gateway/controller/http"
	"github.com/tomocy/go-cookbook/oauth/users/gateway/presentation/json"
	"github.com/tomocy/go-cookbook/oauth/users/infra/memory"
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
		return fmt.Errorf("failed parse args: %w", err)
	}

	var (
		userRepo = memory.NewUserRepo()
	)
	serv := http.NewServer(w, conf.addr, userRepo, json.Renderer)
	if err := serv.Run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
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
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	return nil
}