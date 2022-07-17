FROM alpine:3.11

ARG USERNAME=c8yuser
ARG C8Y_VERSION=1.3.0

RUN apk update \
    && apk add curl unzip bash bash-completion zsh fish git vim jq \
    && adduser -S $USERNAME \
    && mkdir -p /sessions \
    && chown -R $USERNAME /sessions \
    && git clone https://github.com/reubenmiller/go-c8y-cli-addons.git /home/$USERNAME/.go-c8y-cli

WORKDIR /home/$USERNAME

USER $USERNAME
RUN sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)" -s --batch
USER root

# add binary to path
ENV PATH=${PATH}:/home/$USERNAME/bin
ENV C8Y_SESSION_HOME=/sessions
COPY c8y /home/$USERNAME/bin/c8y


# install plugins
RUN echo "source /home/$USERNAME/.go-c8y-cli/shell/c8y.plugin.sh" >> /home/$USERNAME/.bashrc \
    # && echo "export C8Y_SESSION_HOME=/sessions" >> /home/$USERNAME/.bashrc \
    && bash -c "c8y version" \
    #
    # zsh
    && mkdir -p /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/ \
    && cp /home/$USERNAME/.go-c8y-cli/shell/c8y.plugin.zsh /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/ \
    && sed -iE 's/^plugins=(\(.*\))/plugins=(\1 c8y)/' /home/$USERNAME/.zshrc \
    #
    # Create completions before zsh runs otherwise
    # it will not automatically load the completions until the user
    # runs 'source ~/.zshrc'
    && c8y completion zsh > /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/_c8y \
    # && echo "export C8Y_SESSION_HOME=/sessions" >> /home/$USERNAME/.zshrc \
    #
    # fish
    && mkdir -p /home/$USERNAME/.config/fish \
    && echo "source /home/$USERNAME/.go-c8y-cli/shell/c8y.plugin.fish" >> /home/$USERNAME/.config/fish/config.fish \
    # && echo "set -gx C8Y_SESSION_HOME /sessions" >> /home/$USERNAME/.config/fish/config.fish \
    && fish -c "c8y version" \
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
