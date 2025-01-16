FROM golang:alpine
WORKDIR /app
RUN apk add --update --no-cache ffmpeg gcc musl-dev opus-dev pkgconfig py3-pip \ 
    && python3 -m venv ./venv \
    && source ./venv/bin/activate \
    && pip install yt-dlp
RUN rm -rf ~/.cache/pip
COPY . .
RUN go build -ldflags "-s -w"
CMD ["./main"]
EXPOSE ${PORT}