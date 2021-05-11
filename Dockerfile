FROM golang:1.16-alpine3.13 AS build

WORKDIR /app/
ENV GO111MODULE=on

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . /app/
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

FROM scratch

COPY --from=build /app/app /bin/app
EXPOSE 8000
ENTRYPOINT ["/bin/app"]
