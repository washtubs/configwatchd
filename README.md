# `configwatchd`

configwatchd watches changes to a set of config files that you specify,
and executes whatever command you want to restart or trigger a reload
in the corresponding process.

In addition it permits queuing with manual flushing. So instead of immediately
reloading which may be undesirable, the user can reload configs manually.

### Run the server

#### First set up your config

Explicitly list the files you want to watch for each program.
The config is based on the user config directory.
On linux it's `/home/myuser/.config/configwatchd/server.yaml`.

    i3:
      # command is executed by bash
      command: "i3-msg reload"
      watch:
        # tilda (~) expansion is supported (for the beginning of the string)!
        - "~/.i3/config"`

Run the server

    configwatchd serve

### Selective flush with fzf

    configwatchd list | fzf -1 | xargs configwatchd flush

`-1` makes it so if it's just one in the queue you automatically flush it.

Add `-clear` after flush to clear specific ones instead of execute them
