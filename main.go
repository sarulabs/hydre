package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	confFile := "hydre.yml"

	app := cli.NewApp()
	app.Name = "hydre"
	app.Usage = "start several processes in one docker container"
	app.Version = "2.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "conf, c",
			Value:       confFile,
			Usage:       "path to the configuration file",
			Destination: &confFile,
			EnvVar:      "HYDRE_CONFIGURATION_FILE",
		},
	}
	app.Action = func(c *cli.Context) {
		h, err := NewHydre(c.String("conf"))
		if err != nil {
			log.Println(err)
			return
		}
		h.Run()
	}

	app.Run(os.Args)
}
