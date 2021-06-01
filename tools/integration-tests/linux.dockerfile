ARG  IMAGE=debian:10
FROM $IMAGE
WORKDIR /test
COPY install.sh /test/install.sh
RUN chmod +x /test/install.sh
