---
title: odo completion
---

## Description

`odo completion` is used to generate shell completion code. The generated code provides interactive shell completion code for `odo`.

There is support for the following terminal shells:
- [Bash](https://www.gnu.org/software/bash/)
- [Zsh](https://zsh.sourceforge.io/)
- [Fish](https://fishshell.com/)
- [Powershell](https://docs.microsoft.com/en-us/powershell/)

## Running the Command

To generate the shell completion code, the command can be ran as follows:

```sh
odo completion [SHELL]
```

### Bash

Load into your current shell environment:

```sh
source <(odo completion bash)
```

Load persistently:

```sh
# Save the completion to a file
odo completion bash > ~/.odo/completion.bash.inc

# Load the completion from within your $HOME/.bash_profile
source ~/.odo/completion.bash.inc
```

### Zsh

Load into your current shell environment:

```sh
source <(odo zsh)
```

Load persistently:

```sh
odo completion zsh > "${fpath[1]}/_odo"
```

### Fish

Load into your current shell environment:

```sh
source <(odo completion fish)
```

Load persistently:

```sh
odo completion fish > ~/.config/fish/completions/odo.fish
```

### Powershell

Load into your current shell environment:

```sh
odo completion powershell | Out-String | Invoke-Expression
```

Load persistently:

```sh
odo completion powershell >> $PROFILE
```
