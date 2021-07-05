FROM golang:1.16.5-stretch AS build
WORKDIR /build

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build -o redditbot .

FROM golang:1.16.5-stretch AS run
WORKDIR /run
EXPOSE 8080
COPY --from=build /build/redditbot /run

CMD chmod +x ./redditbot

ENTRYPOINT ["./redditbot"]
