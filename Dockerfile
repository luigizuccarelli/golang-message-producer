FROM registry.access.redhat.com/ubi8/ubi-init:latest

# gcc for cgo
RUN dnf install -y git gcc make && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.12.5
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 aea86e3c73495f205929cfebba0d63f1382c8ac59be081b6351681415f4063cf

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
COPY bin/lib* uid_entrypoint.sh microservice /go/

RUN rpm -iv /go/librdkafka-0.11.5-1.el8.x86_64.rpm && rm -rf /go/lib*

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 0755 "$GOPATH"
WORKDIR $GOPATH

USER 1001

ENTRYPOINT [ "./uid_entrypoint.sh" ]

# This will change depending on each microservice entry point
CMD ["./microservice"]
