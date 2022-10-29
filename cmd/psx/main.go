package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bodgit/psx"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

const maxChannels = 8

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var fs = afero.NewOsFs()

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}
}

func saveMemoryCard(base, code string, mc *psx.MemoryCard) error {
	directory := filepath.Join(base, code)
dir:
	fi, err := fs.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			if err := fs.Mkdir(directory, os.ModePerm|os.ModeDir); err != nil {
				return err
			}
			goto dir
		}
		return err
	}
	if !fi.IsDir() {
		return errors.New("not a directory")
	}

	var i int
	var target string
	for i = 1; i <= maxChannels; i++ {
		target = filepath.Join(directory, fmt.Sprintf("%s-%d.mcd", code, i))
		_, err = fs.Stat(target)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			return err
		}
	}

	if i > maxChannels {
		return errors.New("no free memory card channels")
	}

	b, err := mc.MarshalBinary()
	if err != nil {
		return err
	}

	if err := afero.WriteFile(fs, target, b, 0666); err != nil {
		return err
	}

	return nil
}

func sanitizeProductCode(code []byte) []byte {
	clone := make([]byte, len(code))
	copy(clone, code)
	if len(clone) == 10 {
		clone[4] = '-'
	}
	return clone
}

func splitMemoryCard(base string, smc *psx.MemoryCard) error {
	// Create list of unique product codes
	codes := make(map[string]struct{})
	for i := 0; i < psx.NumBlocks; i++ {
		df := smc.HeaderBlock.DirectoryFrame[i]
		sanitized := string(sanitizeProductCode(df.ProductCode[:]))
		if df.AvailableBlocks == psx.BlockFirstLink {
			codes[sanitized] = struct{}{}
		}
	}

	for code := range codes {
		tmc, err := psx.NewMemoryCard()
		if err != nil {
			return err
		}
		i := 0

		for j := 0; j < psx.NumBlocks; j++ {
			df := smc.HeaderBlock.DirectoryFrame[j]
			sanitized := string(sanitizeProductCode(df.ProductCode[:]))
			if df.AvailableBlocks != psx.BlockFirstLink || sanitized != code {
				continue
			}
			for {
				// Copy the directory frame and data block
				tmc.HeaderBlock.DirectoryFrame[i] = df
				if df.LinkOrder != psx.LastLink && tmc.HeaderBlock.DirectoryFrame[i].LinkOrder != uint16(i+1) {
					// Block has moved during the copy
					tmc.HeaderBlock.DirectoryFrame[i].LinkOrder = uint16(i + 1)
					tmc.HeaderBlock.DirectoryFrame[i].UpdateChecksum()
				}
				tmc.DataBlock[i] = smc.DataBlock[j]

				i++

				if df.LinkOrder == psx.LastLink {
					break
				}

				j = int(df.LinkOrder)
				df = smc.HeaderBlock.DirectoryFrame[j]
			}
		}

		if err := saveMemoryCard(base, code, tmc); err != nil {
			return err
		}
	}

	return nil
}

func splitMemoryCards(dir string, files []string) error {
	base, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if fi, err := fs.Stat(base); err != nil || !fi.IsDir() {
		if err != nil {
			return err
		}
		return errors.New("not a directory")
	}

	for _, f := range files {
		source, err := filepath.Abs(f)
		if err != nil {
			return err
		}

		b, err := afero.ReadFile(fs, source)
		if err != nil {
			return err
		}

		mc, err := psx.NewMemoryCard()
		if err != nil {
			return err
		}

		if err := mc.UnmarshalBinary(b); err != nil {
			return err
		}

		if err := splitMemoryCard(base, mc); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := cli.NewApp()

	app.Name = "psx"
	app.Usage = "PlayStation utility"
	app.Version = fmt.Sprintf("%s, commit %s, built at %s", version, commit, date)

	app.Commands = []*cli.Command{
		{
			Name:        "memcardpro",
			Aliases:     []string{"mcp"},
			Usage:       "Manage MemCard PRO virtual memory cards",
			Description: "Manage MemCard PRO virtual memory cards",
			Subcommands: []*cli.Command{
				{
					Name:        "split",
					Usage:       "Split generic virtual memory cards into multiple per-game cards",
					Description: "Split generic virtual memory cards into multiple per-game cards",
					ArgsUsage:   "DIRECTORY FILE...",
					Action: func(c *cli.Context) error {
						if c.NArg() < 2 {
							cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
						}
						return splitMemoryCards(c.Args().First(), c.Args().Tail())
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
