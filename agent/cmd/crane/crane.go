package main

import (
	"context"
	"log"
	"os"

	"github.com/dyrector-io/dyrectorio/agent/internal/util"
	"github.com/dyrector-io/dyrectorio/agent/pkg/crane"
	"github.com/dyrector-io/dyrectorio/agent/pkg/crane/config"
	"github.com/dyrector-io/dyrectorio/agent/pkg/crane/k8sconfig"
	"github.com/dyrector-io/dyrectorio/agent/pkg/version"

	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "crane",
		Version:  version.BuildVersion(),
		HelpName: "crane",
		Usage:    "crane - cli tool for serving a k8s agent of dyrector.io!",
		Action:   serve,

		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Init the key on kubernetes cluster",
				Action:  initKey,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func loadConfiguration() config.Configuration {
	var cfg config.Configuration
	err := util.ReadConfig(&cfg)
	if err != nil {
		log.Panic("failed to load configuration: ", err.Error())
	}
	log.Println("Configuration loaded.")
	return cfg
}

func serve(cCtx *cli.Context) error {
	cfg := loadConfiguration()

	secret, err := k8sconfig.TryToGetValidSecret(context.Background(), cfg.Namespace, cfg.SecretName, &cfg)
	if err != nil {
		return err
	}

	k8sconfig.InjectSecret(secret, &cfg)

	crane.Serve(&cfg)
	return nil
}

func initKey(cCtx *cli.Context) error {
	cfg := loadConfiguration()

	_, err := k8sconfig.TryToGetValidSecret(context.Background(), cfg.Namespace, cfg.SecretName, &cfg)
	if err != nil {
		return err
	}

	return nil
}
