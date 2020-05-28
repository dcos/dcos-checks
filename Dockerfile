FROM golang:1.13

ENV PATH /go/bin:$PATH
ENV GOPATH /go

RUN go get golang.org/x/lint/golint \
	&& go get golang.org/x/tools/cmd/cover \
	&& go install cmd/vet cmd/cover \
    && go get github.com/jstemmer/go-junit-report \
    && go get github.com/smartystreets/goconvey \
    && go get golang.org/x/tools/cmd/cover \
    && go get github.com/axw/gocov/... \
    && go get github.com/AlekSi/gocov-xml


WORKDIR /dcos-checks
