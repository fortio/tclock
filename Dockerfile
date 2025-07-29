FROM scratch
COPY tclock /usr/bin/tclock
ENTRYPOINT ["/usr/bin/tclock"]
