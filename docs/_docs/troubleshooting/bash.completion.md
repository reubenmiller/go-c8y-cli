---
layout: default
category: Troubleshooting
title: Bash Completion Installation
---

## Enabling shell autocompletion

c8y provides autocompletion support for Bash and Zsh, which can save you a lot of typing.

Below are the procedures to set up autocompletion for Bash (including the difference between Linux and macOS) and Zsh.


### Bash on Linux

#### Introduction

The c8y completion script for Bash can be generated with the command c8y completion bash. Sourcing the completion script in your shell enables c8y autocompletion.
However, the completion script depends on [bash-completion](https://github.com/scop/bash-completion), which means that you have to install this software first (you can test if you have bash-completion already installed by running type _init_completion).


#### Install bash-completion

bash-completion is provided by many package managers (see [here](https://github.com/scop/bash-completion#installation)). You can install it with `apt-get install bash-completion` or `yum install bash-completion`, etc.

The above commands create `/usr/share/bash-completion/bash_completion`, which is the main script of bash-completion. Depending on your package manager, you have to manually source this file in your `~/.bashrc` file.

To find out, reload your shell and run `type _init_completion`. If the command succeeds, you’re already set, otherwise add the following to your `~/.bashrc` file:


```sh
source /usr/share/bash-completion/bash_completion
```

Reload your shell and verify that bash-completion is correctly installed with type _init_completion.


### Bash on MacOS

#### Introduction

The c8y completion script for Bash can be generated with c8y completion bash. Sourcing this script in your shell enables c8y completion.
However, the c8y completion script depends on [bash-completion](https://github.com/scop/bash-completion) which you thus have to previously install.

#### Install bash-completion

**Note**
Note: As mentioned, these instructions assume you use Bash 4.1+, which means you will install bash-completion v2 (in contrast to Bash 3.2 and bash-completion v1, in which case c8y completion won’t work).

You can test if you have bash-completion v2 already installed with `type _init_completion`. If not, you can install it with Homebrew:

```sh
brew install bash-completion@2
```

As stated in the output of this command, add the following to your ~/.bashrc file:

```sh
export BASH_COMPLETION_COMPAT_DIR="/usr/local/etc/bash_completion.d"
[[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"
```

Reload your shell and verify that bash-completion v2 is correctly installed with type _init_completion.
