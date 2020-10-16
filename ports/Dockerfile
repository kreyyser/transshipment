FROM golang:latest

COPY ./build_output/ports /usr/local/bin

RUN chmod +x /usr/local/bin/ports

ENTRYPOINT ["/usr/local/bin/ports", "-c"]
CMD ["/usr/local/bin/config.yaml"]