FROM alpine:3.4

RUN mkdir -p /opt/gobot/resources

COPY resources/* /opt/gobot/resources/
COPY gobot /opt/gobot/

WORKDIR /opt/gobot/

CMD ["./gobot"]