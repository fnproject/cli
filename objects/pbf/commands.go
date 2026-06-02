package pbf

import "github.com/urfave/cli"

func sharedListFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{Name: "limit", Usage: "Maximum number of items to return", Value: 10},
		cli.StringFlag{Name: "output", Usage: "Output format (json)"},
	}
}

func List() cli.Command {
	return cli.Command{
		Name:        "pbfs",
		Usage:       "List Pre-Built Function listings and related resources",
		Description: "List available OCI Pre-Built Functions (PBFs), their versions, or supported triggers.",
		Before: func(c *cli.Context) error {
			_, err := initPBFClient()
			return err
		},
		Action: func(c *cli.Context) error {
			cmd, err := initPBFClient()
			if err != nil {
				return err
			}
			return cmd.listListings(c)
		},
		Flags: append(sharedListFlags(),
			cli.StringFlag{Name: "search", Usage: "Filter listings by partial name match"},
			cli.StringFlag{Name: "trigger", Usage: "Filter listings by trigger name"},
		),
		Subcommands: []cli.Command{
			{
				Name:      "versions",
				Usage:     "List versions for a PBF listing",
				ArgsUsage: "<pbf-name-or-ocid>",
				Before: func(c *cli.Context) error {
					_, err := initPBFClient()
					return err
				},
				Action: func(c *cli.Context) error {
					cmd, err := initPBFClient()
					if err != nil {
						return err
					}
					return cmd.listVersions(c)
				},
				Flags: append(sharedListFlags(), cli.BoolFlag{Name: "current", Usage: "Show only the OCI-designated current version"}),
			},
			{
				Name:      "triggers",
				Usage:     "List supported PBF trigger names",
				ArgsUsage: "[trigger-name]",
				Before: func(c *cli.Context) error {
					_, err := initPBFClient()
					return err
				},
				Action: func(c *cli.Context) error {
					cmd, err := initPBFClient()
					if err != nil {
						return err
					}
					return cmd.listTriggers(c)
				},
				Flags: sharedListFlags(),
			},
		},
	}
}

func Get() cli.Command {
	return cli.Command{
		Name:        "pbfs",
		Usage:       "Get a Pre-Built Function listing or version",
		Description: "Get detailed information for an OCI Pre-Built Function (PBF) listing or a specific listing version.",
		ArgsUsage:   "<pbf-name-or-ocid>",
		Before: func(c *cli.Context) error {
			_, err := initPBFClient()
			return err
		},
		Action: func(c *cli.Context) error {
			cmd, err := initPBFClient()
			if err != nil {
				return err
			}
			return cmd.getListing(c)
		},
		Subcommands: []cli.Command{
			{
				Name:      "version",
				Usage:     "Get a specific PBF listing version by OCID",
				ArgsUsage: "<pbf-listing-version-ocid>",
				Before: func(c *cli.Context) error {
					_, err := initPBFClient()
					return err
				},
				Action: func(c *cli.Context) error {
					cmd, err := initPBFClient()
					if err != nil {
						return err
					}
					return cmd.getVersion(c)
				},
			},
		},
	}
}
