
ZSH_THEME="robbyrussell"

plugins=(
    git
    c8y
)

source $ZSH/oh-my-zsh.sh

autoload -U compinit
compinit -i

export LANG=C.UTF-8
export LC_ALL=C.UTF-8
