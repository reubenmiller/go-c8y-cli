FROM ubuntu:22.04
ARG TARGETARCH

ARG USERNAME=c8yuser

# Install powershell as currently there is no official docker image
ENV DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=true
RUN apt-get update && apt-get install -y --no-install-recommends curl wget ca-certificates \
    && case ${TARGETARCH} in \
        "amd64")  PWSH_ARCH=x64  ;; \
        "arm64")  PWSH_ARCH=arm64  ;; \
        "arm64/v8")  PWSH_ARCH=arm64  ;; \
    esac \
    && curl -L -o /tmp/powershell.tar.gz https://github.com/PowerShell/PowerShell/releases/download/v7.4.0/powershell-7.4.0-linux-${PWSH_ARCH}.tar.gz\
    && mkdir -p /opt/microsoft/powershell/7 \
    && tar zxf /tmp/powershell.tar.gz -C /opt/microsoft/powershell/7 \
    && chmod +x /opt/microsoft/powershell/7/pwsh \
    && ln -s /opt/microsoft/powershell/7/pwsh /usr/bin/pwsh

RUN apt-get update \    
    && apt-get install -y vim git jq sudo \
    && adduser --disabled-password --gecos '' $USERNAME \
    && adduser $USERNAME sudo \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/reubenmiller/go-c8y-cli-addons.git /home/${USERNAME}/.go-c8y-cli

ENV C8Y_HOME=/home/${USERNAME}/.go-c8y-cli
ENV C8Y_SESSION_HOME=/sessions
ENV C8Y_INSTALL_OPTIONS="-AllowPrerelease"
COPY output_pwsh/PSc8y /home/${USERNAME}/modules/PSc8y
ENV PSModulePath=/home/${USERNAME}/modules
VOLUME [ "/sessions" ]

COPY docker/profile.ps1 /home/${USERNAME}/.config/powershell/Microsoft.PowerShell_profile.ps1
RUN chown -R $USERNAME /home/$USERNAME

USER $USERNAME

WORKDIR /home/${USERNAME}
ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
