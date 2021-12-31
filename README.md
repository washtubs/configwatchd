# `configwatchd`

### Selective flush with fzf

    configwatchd list | fzf -1 | xargs configwatchd flush

-1 makes it so if it's just one in the queue you automatically flush it.

Add `-clear` after flush to clear specific ones instead of execute them
