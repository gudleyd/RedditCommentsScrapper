FROM golang:1.16.5-stretch AS build
WORKDIR /build
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go build -o redditbot .

FROM golang:1.16.5-stretch AS run
WORKDIR /run
EXPOSE 8080
COPY --from=build /build/redditbot /run
COPY config.json /run
COPY scrapper.agent /run

CMD chmod +x ./redditbot

ENTRYPOINT ["./redditbot"]