package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/glamour"
	goslob "github.com/schoentoon/go-slob"
)

type Application struct {
	cfg         *Config
	slobs       []*goslob.Slob
	files       []*os.File
	suggestions []prompt.Suggest
	lookup      map[string]*goslob.Ref
}

func NewApplication(cfg *Config) (*Application, error) {
	a := &Application{
		cfg:         cfg,
		slobs:       make([]*goslob.Slob, 0, len(cfg.Input)),
		files:       make([]*os.File, 0, len(cfg.Input)),
		suggestions: make([]prompt.Suggest, 0),
		lookup:      make(map[string]*goslob.Ref),
	}

	for _, filename := range cfg.Input {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		a.files = append(a.files, f)

		slob, err := goslob.SlobFromReader(f)
		if err != nil {
			return nil, err
		}
		a.slobs = append(a.slobs, slob)

		ch, _ := slob.Keys()

		// if we have multiple sources, we list the source in the description
		var from string
		if len(cfg.Input) > 1 {
			from = fmt.Sprintf("Source: %s", filename)
		}

		for key := range ch {
			if !a.cfg.SkipKey(key.Key) {
				a.suggestions = append(a.suggestions, prompt.Suggest{Text: key.Key, Description: from})
				a.lookup[key.Key] = key
			}
		}
	}

	return a, nil
}

func (a *Application) completer(d prompt.Document) []prompt.Suggest {
	if len(d.Text) < 3 {
		return nil
	}
	return prompt.FilterHasPrefix(a.suggestions, d.GetWordBeforeCursor(), true)
}

func (a *Application) executor(in string) {
	ref, ok := a.lookup[in]
	if !ok {
		fmt.Printf("Not found\n")
		return
	}

	item, err := ref.Get()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	out, err := ParseItem(item)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	rendered, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	output, err := rendered.Render(out.Render(a.cfg))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("%s", output)
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Printf("Didn't get slob file as an argument\n")
		os.Exit(1)
	}

	cfg, err := ReadConfig(flag.Arg(0))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	a, err := NewApplication(cfg)
	if err != nil {
		log.Fatal(err)
	}

	p := prompt.New(a.executor, a.completer)

	p.Run()
}
