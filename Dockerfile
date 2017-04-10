FROM gcr.io/google-appengine/golang

RUN mkdir -p /opt/gobot/resources

COPY resources/* /opt/gobot/resources/
COPY gobot /opt/gobot/

WORKDIR /opt/gobot/

CMD ["./gobot"]
