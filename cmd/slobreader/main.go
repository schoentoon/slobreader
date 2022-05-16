package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/c-bata/go-prompt"
	goslob "github.com/schoentoon/go-slob"
)

type Application struct {
	Slob *goslob.Slob

	file        *os.File
	suggestions []prompt.Suggest
	lookup      map[string]*goslob.Ref
}

func NewApplication(filename string) (*Application, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	slob, err := goslob.SlobFromReader(f)
	if err != nil {
		return nil, err
	}

	a := &Application{
		Slob:        slob,
		file:        f,
		suggestions: make([]prompt.Suggest, 0),
		lookup:      make(map[string]*goslob.Ref),
	}

	ch, _ := slob.Keys()

	for key := range ch {
		a.suggestions = append(a.suggestions, prompt.Suggest{Text: key.Key})
		a.lookup[key.Key] = key
	}

	return a, nil
}

func (a *Application) completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterFuzzy(a.suggestions, d.GetWordBeforeCursor(), true)
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

	fmt.Printf("%s\n", item.Content)
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Printf("Didn't get slob file as an argument\n")
		os.Exit(1)
	}

	a, err := NewApplication(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	p := prompt.New(a.executor, a.completer)

	p.Run()
}
