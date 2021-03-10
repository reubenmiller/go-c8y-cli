FROM alpine:3.11

ARG USERNAME=c8yuser
ARG C8Y_VERSION=1.3.0

RUN apk update \
    && apk add curl unzip zsh git vim jq \
    && adduser -S $USERNAME \
    && mkdir -p /sessions \
    && chown $USERNAME /sessions

USER $USERNAME
WORKDIR /home/$USERNAME

RUN sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"

# add binary to path
ENV PATH=${PATH}:/home/$USERNAME/bin
ENV C8Y_SESSION_HOME=/sessions

COPY ./c8y.linux /home/$USERNAME/bin/c8y

#
# install c8y zsh plugin
COPY ./c8y.plugin.zsh /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/
RUN sed -iE 's/plugins=(\(.*\))/plugins=(\1 c8y)/' ~/.zshrc \
    #
    # Create completions before zsh runs otherwise
    # it will not automatically load the completions until the user
    # runs 'source ~/.zshrc'
    && c8y completion zsh > /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/_c8y

# Working settings
RUN chown $USERNAME /home/$USERNAME

# Prevent zsh plugins from affecting completions
# https://github.com/ohmyzsh/ohmyzsh/issues/1282
# https://stackoverflow.com/questions/11916064/zsh-tab-completion-duplicating-command-name
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

VOLUME [ "/sessions" ]

ENTRYPOINT [ "/bin/zsh" ]
