FROM mcr.microsoft.com/powershell:7.1.1-ubuntu-18.04

RUN apt-get update \
    && apt-get install -y vim \
    && rm -rf /var/lib/apt/lists/*

ENV C8Y_SESSION_HOME=/sessions
VOLUME [ "/sessions" ]

COPY profile.ps1 /root/.config/powershell/Microsoft.PowerShell_profile.ps1

RUN pwsh -c "Set-PSRepository -Name 'PSGallery' -InstallationPolicy Trusted; " \
    && pwsh -c "Install-Module PSc8y -Repository PSGallery -AllowClobber -Force"

ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
