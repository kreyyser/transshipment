FROM golang:latest

COPY ./build_output/restgateway /usr/local/bin

RUN chmod +x /usr/local/bin/restgateway

ENTRYPOINT ["/usr/local/bin/restgateway", "-c"]
CMD ["/usr/local/bin/config.yaml"]