FROM mcr.microsoft.com/powershell:7.2-ubuntu-20.04

RUN apt-get update \
    && apt-get install -y vim git jq \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/reubenmiller/go-c8y-cli-addons.git /root/.go-c8y-cli

ENV C8Y_HOME=/root/.go-c8y-cli
ENV C8Y_SESSION_HOME=/sessions
ENV C8Y_INSTALL_OPTIONS="-AllowPrerelease"
COPY output_pwsh/PSc8y /root/modules/PSc8y
ENV PSModulePath=/root/modules
VOLUME [ "/sessions" ]

COPY docker/profile.ps1 /root/.config/powershell/Microsoft.PowerShell_profile.ps1

WORKDIR /root
ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
