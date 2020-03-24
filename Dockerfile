FROM golang:1.10.3
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/hcnet/go

COPY . .
RUN dep ensure -v
RUN go install github.com/hcnet/go/tools/...
RUN go install github.com/hcnet/go/services/...
