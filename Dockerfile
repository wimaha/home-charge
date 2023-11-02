FROM golang:1.21-alpine

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /home-charge

EXPOSE 7618

CMD [ "/home-charge" ]