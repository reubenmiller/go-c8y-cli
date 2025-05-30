FROM alpine:3.21

ARG USERNAME=c8yuser

RUN apk update \
    && apk add curl unzip bash bash-completion zsh fish git vim jq sudo coreutils \
    && adduser -S $USERNAME \
    && echo '%wheel ALL=(ALL) ALL' > /etc/sudoers.d/wheel \
    && adduser $USERNAME wheel \
    && mkdir -p /sessions \
    && chown -R $USERNAME /sessions \
    && git clone https://github.com/reubenmiller/go-c8y-cli-addons.git /home/$USERNAME/.go-c8y-cli

WORKDIR /home/$USERNAME

USER $USERNAME
RUN sh -c "$(wget -qO- https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" -s --batch
USER root

# add binary to path
ENV PATH=${PATH}:/home/$USERNAME/bin
ENV C8Y_SESSION_HOME=/sessions
COPY bin/c8y /home/$USERNAME/bin/c8y


# install plugins
RUN sudo -u "${USERNAME}" /home/$USERNAME/bin/c8y cli install \
    && bash -c "c8y version" \
    #
    # zsh
    && zsh -c "c8y version" \
    #
    # fish
    && fish -c "c8y version" \
    #
    # cleanup
    && rm -f /home/$USERNAME/c8y.activitylog*


# Working settings
RUN chown -R $USERNAME /home/$USERNAME

# Prevent zsh plugins from affecting completions
# https://github.com/ohmyzsh/ohmyzsh/issues/1282
# https://stackoverflow.com/questions/11916064/zsh-tab-completion-duplicating-command-name
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

USER $USERNAME
VOLUME [ "/sessions" ]

ENTRYPOINT [ "/bin/zsh" ]
