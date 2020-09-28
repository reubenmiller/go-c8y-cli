FROM alpine:3.11

ARG C8Y_VERSION=1.3.0

RUN apk update \
    && apk add curl unzip zsh git vim jq \
    && adduser -S c8yuser \
    && mkdir -p /sessions \
    && chown c8yuser /sessions

USER c8yuser
WORKDIR /home/c8yuser

RUN sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"

# add binary to path
ENV PATH=${PATH}:/home/c8yuser/bin
ENV C8Y_SESSION_HOME=/sessions

COPY ./c8y.linux /home/c8yuser/bin/c8y

#
# install c8y zsh plugin
COPY ./c8y.plugin.zsh /home/c8yuser/.oh-my-zsh/custom/plugins/c8y/
RUN sed -iE 's/plugins=(\(.*\))/plugins=(\1 c8y)/' ~/.zshrc \
    #
    # Create completions before zsh runs otherwise
    # it will not automatically load the completions until the user
    # runs 'source ~/.zshrc'
    && mkdir -p ~/.oh-my-zsh/completions \
    && c8y completion zsh > ~/.oh-my-zsh/completions/_c8y

# Working settings
RUN chown c8yuser /home/c8yuser

VOLUME [ "/sessions" ]

ENTRYPOINT [ "/bin/zsh" ]
