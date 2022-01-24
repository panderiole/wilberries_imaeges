FROM golang:latest
RUN mkdir /app
ADD ./imagesParser/. /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
