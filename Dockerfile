# docker buildx build . -t germanorizzo/sfup:latest && docker push germanorizzo/sfup:latest

FROM golang:latest as build

WORKDIR /go/src/app
COPY . .

RUN go build -ldflags="-extldflags=-static" -tags sqlite_omit_load_extension -trimpath

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian12
#FROM debian
COPY --from=build /go/src/app/sfup /sfup

EXPOSE 8080
VOLUME /config.yaml
VOLUME /data

ENTRYPOINT ["/sfup", "-config-file", "/config.yaml", "-port", "8080", "-data-dir", "/data"]