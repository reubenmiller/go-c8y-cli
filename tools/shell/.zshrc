
export ZSH=~/.oh-my-zsh
export LANG=C.UTF-8
export LC_ALL=C.UTF-8
export PATH=.bin:$PATH

ZSH_THEME="robbyrussell"

plugins=(
    git
    task
)

if c8y version >/dev/null 2>&1; then
    plugins+=(c8y)
fi

source $ZSH/oh-my-zsh.sh

autoload -U compinit
compinit -i
