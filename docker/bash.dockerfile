FROM alpine:3.11

RUN apk update \
    && apk add curl unzip bash bash-completion \
    && adduser -S c8yuser

WORKDIR /home/c8yuser

RUN curl -L https://www.powershellgallery.com/api/v2/package/PSc8y/1.1.1 -o c8y.zip \
    && unzip -p c8y.zip Dependencies/c8y.linux > /usr/bin/c8y \
    && chmod +x /usr/bin/c8y \
    && rm c8y.zip

USER c8yuser

RUN echo "source /usr/share/bash-completion/bash_completion" >> ~/.bashrc \
    && c8y completion bash >> ~/.bashrc

VOLUME [ "/home/c8yuser/.cumulocity" ]

ENTRYPOINT [ "/bin/bash" ]
