FROM golang:1.14

COPY . /app
WORKDIR /app
RUN go get ./...
ENV GOOS linux
RUN go build -o app
EXPOSE 14000
ENTRYPOINT [ "./app" ]