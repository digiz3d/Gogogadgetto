FROM golang:alpine
RUN apk add --update --no-cache ffmpeg gcc musl-dev opus-dev pkgconfig py3-pip 
RUN pip install --upgrade pip 
RUN pip install yt-dlp
RUN rm -rf ~/.cache/pip

WORKDIR /app
COPY . .
RUN go build -ldflags "-s -w"
CMD ["./main"]
EXPOSE ${PORT}