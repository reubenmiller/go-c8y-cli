FROM mcr.microsoft.com/powershell:7.1.1-ubuntu-18.04

RUN apt-get update \
    && apt-get install -y vim git jq \
    && rm -rf /var/lib/apt/lists/* \
    && git clone https://github.com/reubenmiller/go-c8y-cli-addons.git /root/.go-c8y-cli

ENV C8Y_HOME=/root/.go-c8y-cli
ENV C8Y_SESSION_HOME=/sessions
ENV C8Y_INSTALL_OPTIONS="-AllowPrerelease"
VOLUME [ "/sessions" ]

COPY profile.ps1 /root/.config/powershell/Microsoft.PowerShell_profile.ps1

RUN pwsh -c "Set-PSRepository -Name 'PSGallery' -InstallationPolicy Trusted; " \
    && pwsh -c "Install-Module PSc8y -Repository PSGallery -AllowClobber $C8Y_INSTALL_OPTIONS -Force" \
    && rm -f /root/c8y.activitylog*

ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
