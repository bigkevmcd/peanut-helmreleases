FROM golang:latest AS build
MAINTAINER Kevin McDermott <bigkevmcd@gmail.com>
WORKDIR /go/src
COPY . /go/src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/peanut-pipelines

FROM alpine
WORKDIR /root/
COPY --from=build /go/src/peanut-pipelines .
EXPOSE 8080
ENTRYPOINT ["./peanut-pipelines"]
