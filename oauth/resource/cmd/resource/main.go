package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomocy/go-cookbook/oauth/resource/gateway/controller"
	"github.com/tomocy/go-cookbook/oauth/resource/gateway/presentation"
	"github.com/tomocy/go-cookbook/oauth/resource/infra/memory"
	"github.com/tomocy/go-cookbook/oauth/resource/infra/users"
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
		userServ = users.NewHTTPService(conf.usersAddr)
		userRepo = memory.NewUserRepo()
	)
	ren := presentation.HTML
	ctller := controller.NewHTTPServer(w, conf.addr, ren, userServ, userRepo)
	if err := ctller.Run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

type config struct {
	addr      string
	usersAddr string
}

func (c *config) parse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&c.addr, "addr", ":80", "the address to listen and serve")
	flags.StringVar(&c.usersAddr, "users-addr", "localhost:8080", "the address of users service")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	return nil
}
