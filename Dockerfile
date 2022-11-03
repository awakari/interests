FROM alpine
COPY ./subscriptions /bin/subscriptions
VOLUME ["/etc/subscriptions.yaml"]
ENTRYPOINT ["/bin/subscriptions", "/etc/subscriptions.yaml"]
