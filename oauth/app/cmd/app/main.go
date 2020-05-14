package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tomocy/go-cookbook/oauth"
	"github.com/tomocy/go-cookbook/oauth/app"
	"github.com/tomocy/go-cookbook/oauth/app/gateway/controller"
	"github.com/tomocy/go-cookbook/oauth/app/gateway/presentation"
	"github.com/tomocy/go-cookbook/oauth/app/infra/memory"
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

	oauthClient := oauth.AuthzCodeClient{
		AuthzServerEndpoint: oauth.Endpoint{
			Addr: conf.authzAddr,
			Paths: map[string]string{
				oauth.PathToken: "/tokens",
			},
		},
	}
	ren := presentation.HTML
	userRepo := func() app.UserRepo {
		repo := memory.NewUserRepo()
		u, _ := app.NewUser("aiueo_user_id")
		repo.Save(context.Background(), u)

		return repo
	}()
	con := controller.NewHTTPServer(w, conf.addr, oauthClient, ren, userRepo)
	if err := con.Run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

type config struct {
	addr      string
	authzAddr string
}

func (c *config) parse(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.StringVar(&c.addr, "addr", ":80", "the address to listen and serve")
	flags.StringVar(&c.authzAddr, "authz-addr", "localhost:8080", "the address of authorization service")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	return nil
}
