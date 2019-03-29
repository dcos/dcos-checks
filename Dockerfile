FROM docker:dind

ENV PATH /go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	bash \
	ca-certificates \
	curl \
	make \
	gcc \
	go \
	git \
	libc-dev \
	libgcc \
	make \
	diffutils \
	jq \
	file

RUN go get github.com/golang/lint/golint \
	&& go get golang.org/x/tools/cmd/cover \
	&& go install cmd/vet cmd/cover \
    && go get -u github.com/jstemmer/go-junit-report \
    && go get -u github.com/smartystreets/goconvey \
    && go get -u golang.org/x/tools/cmd/cover \
    && go get -u github.com/axw/gocov/... \
    && go get -u github.com/AlekSi/gocov-xml


WORKDIR /dcos-checks
