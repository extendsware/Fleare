FROM debian

WORKDIR /app

ARG APP_VERSION

COPY releases/linux-amd64/fleare-$APP_VERSION-linux-amd64.tar.gz ./

COPY install-docker.sh ./

RUN tar -xvf fleare-$APP_VERSION-linux-amd64.tar.gz

RUN chmod +x /app/install-docker.sh

RUN /app/install-docker.sh

RUN rm /app/install-docker.sh \
    /app/fleare-$APP_VERSION-linux-amd64.tar.gz \
    fleare-cli \
    fleare-$APP_VERSION-linux-amd64  \
    install_fleare.sh

EXPOSE 4775

CMD ["fleare"]
