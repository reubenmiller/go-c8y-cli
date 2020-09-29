FROM mcr.microsoft.com/powershell

ENV C8Y_SESSION_HOME=/sessions
VOLUME [ "/sessions" ]

RUN pwsh -c "Set-PSRepository -Name 'PSGallery' -InstallationPolicy Trusted; " \
    && pwsh -c "Install-Module PSc8y -Repository PSGallery -AllowClobber -Force"

ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
