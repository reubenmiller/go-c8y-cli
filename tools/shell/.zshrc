
export ZSH=~/.oh-my-zsh
export LANG=C.UTF-8
export LC_ALL=C.UTF-8
export PATH=$PWD/.bin:$PATH

ZSH_THEME="robbyrussell"

plugins=(
    git
    task
)

if c8y version >/dev/null 2>&1; then
    eval "$(c8y cli profile)"
fi

source $ZSH/oh-my-zsh.sh

autoload -U compinit
compinit -i
