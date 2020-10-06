FROM debian:buster
LABEL maintainer="Charlie Drage <charlie@charliedrage.com>"

RUN apt-get update && apt-get install -y \
      asciidoctor \
      pandoc

RUN useradd -ms /bin/bash user
USER user
WORKDIR /home/user
