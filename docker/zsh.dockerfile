FROM alpine:3.11

RUN apk update \
    && apk add curl unzip zsh git \
    && adduser -S c8yuser

RUN curl -L https://www.powershellgallery.com/api/v2/package/PSc8y/1.2.0 -o c8y.zip \
    && unzip -p c8y.zip Dependencies/c8y.linux > /usr/bin/c8y \
    && chmod +x /usr/bin/c8y \
    && rm c8y.zip

USER c8yuser
WORKDIR /home/c8yuser

RUN sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"

# Add completions to the fpath. It must be called _c8y!
RUN mkdir ~/.oh-my-zsh/completions \
    && c8y completion zsh >> ~/.oh-my-zsh/completions/_c8y

VOLUME [ "/home/c8yuser/.cumulocity" ]

ENTRYPOINT [ "/bin/zsh" ]
