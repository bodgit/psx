package main

import (
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
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

var (
	errNoFreeChannels = errors.New("no free memory card channels")
	errNotDirectory   = errors.New("not a directory")
)

func init() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}
}

func newMemoryCardFile(base, code string) (afero.File, error) {
	directory := filepath.Join(base, code)
dir:
	fi, err := fs.Stat(directory)

	if err != nil {
		if os.IsNotExist(err) {
			if err := fs.Mkdir(directory, os.ModePerm|os.ModeDir); err != nil {
				return nil, fmt.Errorf("unable to create directory: %w", err)
			}

			goto dir
		}

		return nil, fmt.Errorf("unable to stat directory: %w", err)
	}

	if !fi.IsDir() {
		return nil, errNotDirectory
	}

	var (
		i      int
		target string
	)

	for i = 1; i <= maxChannels; i++ {
		target = filepath.Join(directory, fmt.Sprintf("%s-%d.mcd", code, i))
		if _, err = fs.Stat(target); err != nil {
			if os.IsNotExist(err) {
				break
			}

			return nil, fmt.Errorf("unable to stat directory: %w", err)
		}
	}

	if i > maxChannels {
		return nil, errNoFreeChannels
	}

	file, err := fs.Create(target)
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %w", err)
	}

	return file, nil
}

func sanitizeProductCode(code string) string {
	if code[4] == 'P' {
		return code[:4] + "-" + code[5:]
	}

	return code
}

type opener interface {
	Open() (iofs.File, error)
}

type writer interface {
	Create() (io.WriteCloser, error)
}

func copyData(f opener, w writer) error {
	rc, err := f.Open()
	if err != nil {
		return err //nolint:wrapcheck
	}
	defer rc.Close()

	wc, err := w.Create()
	if err != nil {
		return err //nolint:wrapcheck
	}

	if _, err := io.Copy(wc, rc); err != nil {
		return fmt.Errorf("unable to copy: %w", err)
	}

	return wc.Close() //nolint:wrapcheck
}

func splitMemoryCard(base string, r *psx.Reader) error {
	// Create list of unique product codes
	codes := make(map[string]struct{})
	for _, file := range r.File {
		codes[sanitizeProductCode(file.ProductCode)] = struct{}{}
	}

	for code := range codes {
		f, err := newMemoryCardFile(base, code)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := psx.NewWriter(f)
		if err != nil {
			return err //nolint:wrapcheck
		}
		defer w.Close()

		for _, file := range r.File {
			if sanitizeProductCode(file.ProductCode) != code {
				continue
			}

			if err := copyData(file, w); err != nil {
				return err
			}
		}
	}

	return nil
}

func splitMemoryCards(dir string, files []string) error {
	base, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("unable to create absolute path: %w", err)
	}

	if fi, err := fs.Stat(base); err != nil || !fi.IsDir() {
		if err != nil {
			return fmt.Errorf("unable to stat directory: %w", err)
		}

		return errNotDirectory
	}

	for _, f := range files {
		source, err := filepath.Abs(f)
		if err != nil {
			return fmt.Errorf("unable to create absolute path: %w", err)
		}

		file, err := fs.Open(source)
		if err != nil {
			return fmt.Errorf("unable to open: %w", err)
		}
		defer file.Close()

		r, err := psx.NewReader(file)
		if err != nil {
			return err //nolint:wrapcheck
		}

		if err := splitMemoryCard(base, r); err != nil {
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
						if c.NArg() < 2 { //nolint:gomnd
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
