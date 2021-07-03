FROM golang:1.16.5-stretch AS build
WORKDIR /build
COPY src/ /build/src
COPY go.mod /build
COPY go.sum /build
RUN go build -o redditbot /build/src

FROM golang:1.16.5-stretch AS run
WORKDIR /run
EXPOSE 8080
COPY --from=build /build/redditbot /run
COPY config.json /run
COPY scrapper.agent /run/scrapper.agent

CMD chmod +x ./redditbot

ENTRYPOINT ["./redditbot"]