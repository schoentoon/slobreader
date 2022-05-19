package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/adrg/xdg"
	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/glamour"
	"github.com/mitchellh/go-homedir"
	goslob "github.com/schoentoon/go-slob"
)

type Application struct {
	cfg   *Config
	slobs []*goslob.Slob
	files []*os.File

	ready       int32
	mutex       sync.RWMutex
	suggestions []prompt.Suggest
}

func NewApplication(cfg *Config) (*Application, error) {
	a := &Application{
		cfg:         cfg,
		slobs:       make([]*goslob.Slob, 0, len(cfg.Input)),
		files:       make([]*os.File, 0, len(cfg.Input)),
		ready:       0,
		suggestions: make([]prompt.Suggest, 0),
	}

	for _, filename := range cfg.Input {
		fullpath, err := homedir.Expand(filename)
		if err != nil {
			return nil, err
		}

		f, err := os.Open(fullpath)
		if err != nil {
			return nil, err
		}
		a.files = append(a.files, f)

		slob, err := goslob.SlobFromReader(f)
		if err != nil {
			return nil, err
		}
		a.slobs = append(a.slobs, slob)

		if !cfg.Autocomplete.Disable {
			go a.fillSuggestions(filename, slob)
		}
	}

	return a, nil
}

func (a *Application) fillSuggestions(filename string, slob *goslob.Slob) {
	ch, _ := slob.Keys()
	suggestions := make([]prompt.Suggest, 0)

	// if we have multiple sources, we list the source in the description
	var from string
	if len(a.cfg.Input) > 1 {
		from = fmt.Sprintf("Source: %s", filename)
	}

	for key := range ch {
		if !a.cfg.SkipKey(key.Key) {
			suggestions = append(suggestions, prompt.Suggest{Text: key.Key, Description: from})
		}
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.suggestions = append(a.suggestions, suggestions...)

	atomic.AddInt32(&a.ready, 1)
}

func (a *Application) completer(d prompt.Document) []prompt.Suggest {
	if a.cfg.Autocomplete.Disable {
		return nil
	}

	if len(d.Text) < 3 {
		return nil
	}

	if atomic.LoadInt32(&a.ready) != int32(len(a.cfg.Input)) {
		return nil
	}

	return prompt.FilterHasPrefix(a.suggestions, d.GetWordBeforeCursor(), true)
}

func (a *Application) executor(in string) {
	in = strings.Trim(in, " ")
	if in == "" { // we ignore empty input
		return
	}

	var item *goslob.Item
	var err error

	for _, slob := range a.slobs {
		item, err = slob.Find(in)
		if err == nil {
			break
		}
	}

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

	cfgfile, _ := xdg.ConfigFile("slobreader/default.yml")
	if flag.NArg() > 0 {
		cfgfile = flag.Arg(0)
	}

	cfgfile, err := homedir.Expand(cfgfile)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	cfg, err := ReadConfig(cfgfile)
	if err != nil {
		fmt.Printf("%s, opening as a slob file instead.\n", err)

		cfg = &Config{
			Input: []string{cfgfile},
		}
	}

	a, err := NewApplication(cfg)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	p := prompt.New(a.executor, a.completer)

	p.Run()
}
