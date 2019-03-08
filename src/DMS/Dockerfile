FROM golang:latest
RUN mkdir -p /home/hdx/
RUN cd /home/hdx/
ADD main /home/hdx/
ADD config.yml /go/
RUN chmod +x /home/hdx/main
ENTRYPOINT ["/home/hdx/main"]