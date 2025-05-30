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
ENV PATH=${PATH}:/home/$USERNAME/.bin
ENV C8Y_SESSION_HOME=/sessions
COPY bin/c8y /home/$USERNAME/.bin/c8y
COPY docker/zshrc /home/${USERNAME}/.zshrc

# Copy zsh profile as running 'c8y cli install' does not work at build time as there an incompatibility with emulation
# see https://github.com/tonistiigi/binfmt/issues/245
COPY pkg/cmd/cli/profile/scripts/plugin.sh /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/c8y.plugin.zsh
COPY output/zsh/_c8y /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/_c8y


# install plugins
RUN echo "installing c8y shell profiles" \
    # allow sudo usage
    && echo "$USERNAME ALL=(ALL:ALL) NOPASSWD: ALL" | sudo tee "/etc/sudoers.d/dont-prompt-$USERNAME-for-sudo-password" \
    && rm -f /etc/sudoers.d/wheel \
    #
    # bash
    && mkdir -p "/home/$USERNAME/.bash_completion.d" \
    && wget -O - https://raw.githubusercontent.com/cykerway/complete-alias/master/complete_alias > "/home/$USERNAME/.bash_completion.d/complete_alias" \
    && printf 'eval "$(c8y cli profile --shell bash)"' >> /home/$USERNAME/.bashrc \
    #
    # zsh
    # no custom operations are needed, as files are already copied in a previous step
    #
    # fish
    && mkdir -p /home/$USERNAME/.config/fish \
    && echo "c8y cli profile --shell fish | source" >> /home/$USERNAME/.config/fish/config.fish \
    #
    # Cleanup
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

# Use CMD over ENTRYPOINT to allow user to re-use the image easily for different purposes
CMD [ "/bin/zsh" ]
