FROM alpine:3.11

ARG C8Y_VERSION=1.3.0

RUN apk update \
    && apk add curl unzip bash bash-completion vim jq \
    && adduser -S c8yuser \
    && mkdir -p /sessions \
    && chown c8yuser /sessions

USER c8yuser
WORKDIR /home/c8yuser

COPY ./c8y.linux /home/c8yuser/bin/c8y
COPY ./c8y.plugin.sh /home/c8yuser/

ENV PATH=${PATH}:/home/c8yuser/bin
ENV C8Y_SESSION_HOME=/sessions

RUN echo "source /home/c8yuser/c8y.plugin.sh" >> /home/c8yuser/.bashrc \
    && bash -c "source ~/c8y.plugin.sh; c8y version"

VOLUME [ "/sessions" ]

ENTRYPOINT [ "/bin/bash" ]
