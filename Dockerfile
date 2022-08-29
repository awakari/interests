FROM alpine
COPY ./subscriptions /bin/subscriptions
COPY ./version.txt /tmp/version.txt
COPY ./glob/globs.csv /tmp/globs.csv
RUN \
    export VERSION=$(cat /tmp/version.txt) && \
    mkdir -p ${HOME}/.subscriptions/${VERSION} && \
    rm -f /tmp/version.txt
ENTRYPOINT ["/bin/subscriptions"]
