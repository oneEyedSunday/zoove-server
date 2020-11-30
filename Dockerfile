FROM golang:1.14

COPY . /app
WORKDIR /app
RUN go get ./
ENV PORT=8300
ENV GOOS linux
RUN go run github.com/prisma/prisma-client-go prefetch
RUN go run github.com/prisma/prisma-client-go generate
RUN go run github.com/prisma/prisma-client-go migrate up --experimental
RUN go run github.com/prisma/prisma-client-go migrate save --experimental --create-db --name "prod"
RUN go build -o app
EXPOSE 8300
ENTRYPOINT [ "./app" ]