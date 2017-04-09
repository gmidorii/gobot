FROM golang:1.8.1-alpine

RUN mkdir -p /opt/gobot/resources

COPY resources/* /opt/gobot/resources/
COPY gobot /opt/gobot/

WORKDIR /opt/gobot/

CMD ["./gobot"]
