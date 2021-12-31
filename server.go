package configwatchd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type ServerOptions struct {
	// Enables an in-memory queue of processes that should be reloaded
	Queue bool
	// Print debug information to stderr
	Verbose bool
	// Location of the config file
	ConfigFile string
}

type MainConfig map[string]Config

type Config struct {
	Command string   `yaml:"command"`
	Watch   []string `yaml:"watch"`
}

func execute(configKey string, mainConfig MainConfig) {
	// TODO: verbose only
	log.Printf("executing %s", configKey)
	cmd := exec.Command("bash", "-c", mainConfig[configKey].Command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command [%s] exited with non-zero code: %s\nOutput:\n%s",
			mainConfig[configKey].Command, err.Error(), string(out))
	}
}

func loadConfig(configFile string) (MainConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to open %s: %w", configFile, err)
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	mainConfig := new(MainConfig)
	decoder := yaml.NewDecoder(bytes.NewReader(bs))
	err = decoder.Decode(mainConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode yaml from %s: %w", configFile, err)
	}

	return *mainConfig, nil

}

func convertWatch(path string) (string, error) {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	return filepath.EvalSymlinks(path)
}

// Starts the server
// + Listens on RPC for queue requests
// + Sets up an fsnotify watcher for the files of each process
func StartServer(opts ServerOptions) error {
	if opts.ConfigFile == "" {
		return errors.New("Must provide a config file")
	}
	mainConfig, err := loadConfig(opts.ConfigFile)
	if err != nil {
		return err
	}

	q := queue{}
	q.contents = make([]string, 0, 10)

	l, err := setupReceiver(&q, opts)
	if err != nil {
		return fmt.Errorf("Failed to setup RPC receiver: %w", err)
	}
	defer l.Close()

	for configKey := range mainConfig {
		log.Print(configKey)
	}
	for configKey, config := range mainConfig {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("Failed to create fsnotify watcher: %w", err)
		}

		for _, watch := range config.Watch {
			watchConverted, err := convertWatch(watch)
			log.Printf("watching %s", watchConverted)
			if err != nil {
				log.Printf("Error resolving %s: %s", watch, err)
				continue
			}

			err = watcher.Add(watchConverted)
			if err != nil {
				log.Printf("Error adding watcher for %s: %s", watch, err.Error())
			}
		}
		go func(configKey string) {
			for {
				select {
				case ev, ok := <-watcher.Events:
					if !ok {
						break
					}
					if ev.Op != fsnotify.Write {
						break
					}
					// TODO: verbose only
					log.Printf("[%s] op=%s name=%s", configKey, ev.Op, ev.Name)
					if opts.Queue {
						q.enqueue(configKey)
					} else {
						execute(configKey, mainConfig)
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						break
					}
					log.Printf("Watch error occured %s", err.Error())
				}
			}
		}(configKey)
	}

	for {
		// Sleep indefinitely until we get an interrupt
		time.Sleep(10 * time.Second)
	}
}
