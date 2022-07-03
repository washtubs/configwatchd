package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/washtubs/configwatchd"
)

var serveFs *flag.FlagSet = flag.NewFlagSet("serve", flag.ExitOnError)
var flushFs *flag.FlagSet = flag.NewFlagSet("flush", flag.ExitOnError)
var listFs *flag.FlagSet = flag.NewFlagSet("list", flag.ExitOnError)

var flushOpts configwatchd.FlushOpts = configwatchd.FlushOpts{}
var serverOpts configwatchd.ServerOptions = configwatchd.ServerOptions{}

func serve(args []string) {
	err := serveFs.Parse(args)
	if err != nil {
		log.Fatal(err)
	}

	var debug *log.Logger
	if serverOpts.Verbose {
		debug = log.Default()
	} else {
		debug = log.New(io.Discard, "", 0)
	}
	configwatchd.SetLoggers(struct {
		Debug *log.Logger
		Error *log.Logger
	}{
		Debug: debug,
		Error: log.Default(),
	})

	err = configwatchd.StartServer(serverOpts)
	if err != nil {
		log.Fatal(err)
	}

}

func flush(args []string) {
	err := flushFs.Parse(args)
	if err != nil {
		log.Fatal(err)
	}
	flushOpts.Keys = flushFs.Args()
	err = configwatchd.Flush(flushOpts)
	if err != nil {
		log.Fatal(err)
	}

}

func list(args []string) {
	list, err := configwatchd.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range list {
		fmt.Println(v)
	}

}

func usage(fs *flag.FlagSet, form string, description string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "%s %s\n  %s\n", fs.Name(), form, description)
		fs.PrintDefaults()
	}
}

func usageFull() {
	description := `
configwatchd watches changes to a set of config files that you specify,
and executes whatever command you want to restart or trigger a reload
in the corresponding process.

In addition it permits queuing with manual flushing. So instead of immediately
reloading which may be undesirable, the user can reload configs manually.

The config file is yaml and of the following form:

  i3:
    # command is executed by bash
    command: "i3-msg reload"
    watch:
      # tilda (~) expansion is supported (for the beginning of the string)!
      - "~/.i3/config"`
	os.Stderr.WriteString("configwatchd (serve|flush|list) [OPTION]\n" + description + "\n\n")
	serveFs.Usage()
	os.Stderr.WriteString("\n")
	flushFs.Usage()
	os.Stderr.WriteString("\n")
	listFs.Usage()
}

func main() {
	userConfigDir, err := os.UserConfigDir()
	defaultServerConfigPath := path.Join(userConfigDir, "configwatchd", "server.yaml")
	if err != nil {
		defaultServerConfigPath = ""
	}

	serveFs.BoolVar(&serverOpts.Queue, "queue", false, "Instead of execute commands as soon as files change, queue them to be executed manually by flush")
	serveFs.BoolVar(&serverOpts.Verbose, "v", false, "Print debug information to stderr")
	serveFs.StringVar(&serverOpts.ConfigFile, "config-file", defaultServerConfigPath, "Override the path to server config")
	flushFs.BoolVar(&flushOpts.Clear, "clear", false, "Instead of execute, clear")
	serveFs.Usage = usage(serveFs, "[OPTION]", "Runs the file watcher / server.")
	flushFs.Usage = usage(flushFs, "[OPTION] [...KEYS]", "Tells the server to flush the queue. Optionally pass specific keys you want to process.")
	listFs.Usage = usage(listFs, "", "Gets the contents of the queue from the server, and prints to stdout.")

	if len(os.Args) < 2 {
		usageFull()
		os.Exit(1)
	}

	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	} else {
		args = []string{}
	}
	switch os.Args[1] {
	case "help":
		usageFull()
	case "serve":
		serve(args)
	case "flush":
		flush(args)
	case "list":
		list(args)
	default:
		usageFull()
		os.Exit(1)
	}
}
