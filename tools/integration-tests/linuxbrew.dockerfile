ARG  IMAGE=ubuntu:latest
FROM $IMAGE

ENV CI=true
RUN apt-get update \
    && apt-get install -y git bash sudo curl procps gcc \
    && /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

WORKDIR /test
COPY install.sh /test/install.homebrew.sh
RUN chmod +x /test/install.homebrew.sh
