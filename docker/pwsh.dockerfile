FROM mcr.microsoft.com/powershell

RUN pwsh -c "Install-Module PSc8y -Repository PSGallery -AllowPrerelease -AllowClobber -Force"

VOLUME [ "/root/.cumulocity" ]

ENTRYPOINT [ "pwsh", "-NoExit", "-c", "Import-Module PSc8y" ]
