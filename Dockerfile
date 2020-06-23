FROM golang:1.14 as builder
RUN adduser --system appuser

WORKDIR $GOPATH/src/gitlab.switch.ch/memoriav/memobase-2020/services/histogram
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app -a gitlab.switch.ch/memoriav/memobase-2020/services/histogram/cmd/webservice

FROM perl:5.30-slim-buster
WORKDIR /app
RUN groupadd -r appuser && \
chmod -R 770 /colormap && \
chown -R :appuser /colormap
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/gitlab.switch.ch/memoriav/memobase-2020/services/histogram/bin/app /app
COPY --from=builder /etc/passwd /etc/passwd

RUN apt-get update && \
apt-get install -y exiftool ffmpeg imagemagick && \
apt-get autoremove -y && \
apt-get clean
# ADD ffprobe /usr/bin/
# ADD convert /usr/bin/
# ADD identify /usr/bin/

USER appuser:appuser

EXPOSE 8083

ENTRYPOINT ["/app/app"]
