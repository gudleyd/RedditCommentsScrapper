FROM golang:1.16.5-stretch AS build
WORKDIR /build
COPY config /build/config
COPY dbmanager /build/dbmanager
COPY handlers /build/handlers
COPY logic /build/logic
COPY models /build/models
COPY mylog /build/mylog
COPY observer /build/observer
COPY main.go /build
COPY go.mod /build
COPY go.sum /build
RUN go build -o redditbot .

FROM golang:1.16.5-stretch AS run
WORKDIR /run
EXPOSE 8080
COPY --from=build /build/redditbot /run

CMD chmod +x ./redditbot

ENTRYPOINT ["./redditbot"]
