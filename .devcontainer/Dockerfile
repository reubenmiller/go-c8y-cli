#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

FROM golang:1.16

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

# This Dockerfile adds a non-root user with sudo access. Use the "remoteUser"
# property in devcontainer.json to use it. On Linux, the container user's GID/UIDs
# will be updated to match your local UID/GID (when using the dockerFile property).
# See https://aka.ms/vscode-remote/containers/non-root-user for details.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

# Configure apt, install packages and tools
RUN wget https://packages.microsoft.com/config/debian/10/packages-microsoft-prod.deb \
    && dpkg -i packages-microsoft-prod.deb \
    && curl -sL https://deb.nodesource.com/setup_14.x | bash - \
    #
    # Update and install packages
    #
    && apt-get update \
    #
    && apt-get -y install --no-install-recommends apt-utils dialog 2>&1 \
    #
    # Verify git, process tools, lsb-release (common in install instructions for CLIs) installed
    && apt-get -y install git openssh-client less iproute2 procps lsb-release gnupg \
    #
    # Extra tooling
    && apt-get -y install make powershell nodejs unzip lolcat boxes figlet randtype \
    && npm install -g yamljs svg-term-cli \
    #
    # Build Go tools w/module support
    && mkdir -p /tmp/gotools \
    && cd /tmp/gotools \
    && GOPATH=/tmp/gotools GO111MODULE=on go get -v golang.org/x/tools/gopls@latest 2>&1 \
    && GOPATH=/tmp/gotools GO111MODULE=on go get -v \
        honnef.co/go/tools/...@latest \
        golang.org/x/tools/cmd/gorename@latest \
        golang.org/x/tools/cmd/goimports@latest \
        golang.org/x/tools/cmd/guru@latest \
        golang.org/x/lint/golint@latest \
        github.com/mdempsky/gocode@latest \
        github.com/cweill/gotests/...@latest \
        github.com/haya14busa/goplay/cmd/goplay@latest \
        github.com/sqs/goreturns@latest \
        github.com/josharian/impl@latest \
        github.com/davidrjenni/reftools/cmd/fillstruct@latest \
        github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest  \
        github.com/ramya-rao-a/go-outline@latest  \
        github.com/acroca/go-symbols@latest  \
        github.com/godoctor/godoctor@latest  \
        github.com/rogpeppe/godef@latest  \
        github.com/zmb3/gogetdoc@latest \
        github.com/fatih/gomodifytags@latest  \
        github.com/mgechev/revive@latest  \
        github.com/go-delve/delve/cmd/dlv@latest 2>&1 \
    #
    # Install Go tools
    && mv /tmp/gotools/bin/* /usr/local/bin/ \
    #
    # Install golangci-lint
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/local/bin 2>&1 \
    #
    # Create a non-root user to use if preferred - see https://aka.ms/vscode-remote/containers/non-root-user.
    && groupadd --gid $USER_GID $USERNAME \
    && useradd -s /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME \
    # [Optional] Add sudo support
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    #
    # Install Docker CE cli
    && apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common lsb-release \
    && curl -fsSL https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]')/gpg | apt-key add - 2>/dev/null \
    && add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') $(lsb_release -cs) stable" \
    && apt-get update \
    && apt-get install -y docker-ce-cli \
    #
    # Install Docker Compose
    && curl -sSL "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose \
    #
    # Install bash tools
    # https://github.com/bats-core/bats-core
    && npm install -g bats \
    && apt-get install -y jq bash-completion asciinema \
    #
    # Fish shell
    && apt-get install -y fish zsh \
    #
    # Clean up
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /tmp/gotools

#
# Install password manager/s
#
RUN curl -sSfL "https://vault.bitwarden.com/download/?app=cli&platform=linux" -o /tmp/bw.zip \
    && unzip /tmp/bw.zip -d /usr/local/bin/ \
    && chmod a+x /usr/local/bin/bw \
    && rm -f /tmp/bw.zip

#
# Shells
#
RUN echo "Installing shells" \
    # Fish shell
    && apt-get install -y fish zsh


# AsciiCinema configuration
RUN chown -R $USER_UID:$USER_GID /home/$USERNAME \
    && mkdir -p /home/$USERNAME/.config/asciinema \
COPY tools/asciinema/config /home/$USERNAME/.config/asciinema/

# Install powershell dependencies
USER $USERNAME
#RUN pwsh -Command "Install-Module -Name platyPS -Scope CurrentUser -Force" \
#   && pwsh -Command "Install-Module -Name Pester -Scope CurrentUser -MinimumVersion '5.0.0' -Force"
    #
    # create symlink to c8y
#    && mkdir -p /workspaces/go-c8y-cli/tools/PSc8y/dist/PSc8y/Dependencies/ \
#    && ln -s /workspaces/go-c8y-cli/tools/PSc8y/dist/PSc8y/Dependencies/c8y.linux /usr/local/bin/c8y

# Install oh-my-zsh
RUN curl -sSfL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh | sh -s -- "" --unattended 2>&1 \
    && mkdir -p /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y

COPY tools/shell/.zshrc /home/$USERNAME/.zshrc
COPY tools/shell/c8y.plugin.zsh /home/$USERNAME/.oh-my-zsh/custom/plugins/c8y/
COPY tools/shell/c8y.plugin.fish /home/$USERNAME/
RUN sed -iE 's/plugins=(\(.*\))/plugins=(\1 c8y)/' /home/$USERNAME/.zshrc \
    && echo "autoload -U compinit; compinit -i" >> /home/$USERNAME/.zshrc \
    #
    # fish config
    && sudo mkdir -p /home/$USERNAME/.config/fish/
    # && echo "c8y completion fish | source" >> /home/$USERNAME/.config/fish/config.fish
    # && echo "source ~/c8y.plugin.fish" >> /home/$USERNAME/.config/fish/config.fish

# CI/CD tools
RUN  cd /tmp \
    && curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh \
    && chmod a+x /tmp/bin/* \
    && sudo mv /tmp/bin/* /usr/local/bin/ \
    && curl -fL https://getcli.jfrog.io | sh \
    && chmod a+x jfrog \
    && sudo mv jfrog /usr/local/bin/

# Update this to "on" or "off" as appropriate
ENV GO111MODULE=auto
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Change default location of cumulocity profiles
ENV C8Y_SESSION_HOME=/workspaces/go-c8y-cli/.cumulocity

# Switch back to dialog for any ad-hoc use of apt-get
ENV DEBIAN_FRONTEND=dialog
