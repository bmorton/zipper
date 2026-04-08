package main

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/bmorton/zipper"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "zipper",
		Usage: "US zip code lookup utility",
		Commands: []*cli.Command{
			{
				Name:      "lookup",
				Usage:     "Look up a zip code",
				ArgsUsage: "<zipcode>",
				Action:    lookupAction,
			},
			{
				Name:      "nearest",
				Usage:     "Find the nearest zip code(s) to coordinates",
				ArgsUsage: "<lat> <lon>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "count",
						Aliases: []string{"n"},
						Value:   5,
						Usage:   "number of results to return",
					},
				},
				Action: nearestAction,
			},
			{
				Name:      "within",
				Usage:     "Find all zip codes within a radius of coordinates",
				ArgsUsage: "<lat> <lon> <radius_km>",
				Action:    withinAction,
			},
			{
				Name:   "info",
				Usage:  "Show dataset info",
				Action: infoAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func lookupAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return cli.ShowCommandHelp(c, "lookup")
	}

	zip := c.Args().First()
	entries := zipper.Lookup(zip)
	if entries == nil {
		fmt.Printf("No results for zip code %s\n", zip)
		return nil
	}

	printEntries(entries)
	return nil
}

func nearestAction(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.ShowCommandHelp(c, "nearest")
	}

	lat, err := strconv.ParseFloat(c.Args().Get(0), 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(c.Args().Get(1), 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}

	n := c.Int("count")
	entries := zipper.NearestN(lat, lon, n)
	if len(entries) == 0 {
		fmt.Println("No results found")
		return nil
	}

	printEntries(entries)
	return nil
}

func withinAction(c *cli.Context) error {
	if c.NArg() < 3 {
		return cli.ShowCommandHelp(c, "within")
	}

	lat, err := strconv.ParseFloat(c.Args().Get(0), 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(c.Args().Get(1), 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}
	radius, err := strconv.ParseFloat(c.Args().Get(2), 64)
	if err != nil {
		return fmt.Errorf("invalid radius: %w", err)
	}

	entries := zipper.WithinRadius(lat, lon, radius)
	if len(entries) == 0 {
		fmt.Println("No results within radius")
		return nil
	}

	fmt.Printf("Found %d zip codes within %.1f km\n\n", len(entries), radius)
	printEntries(entries)
	return nil
}

func infoAction(c *cli.Context) error {
	fmt.Printf("Data version: %s\n", zipper.DataVersion())
	fmt.Printf("Total entries: %d\n", zipper.Size())
	return nil
}

func printEntries(entries []zipper.Entry) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ZIP\tCITY\tSTATE\tLAT\tLON")
	_, _ = fmt.Fprintln(w, "-----\t----\t-----\t---\t---")
	for _, e := range entries {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s (%s)\t%.4f\t%.4f\n",
			e.ZipCode, e.City, e.State, e.StateCode, e.Lat, e.Lon)
	}
	_ = w.Flush()
}
