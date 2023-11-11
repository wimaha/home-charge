############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
#ENV CGO_ENABLED=1
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git && apk add build-base && apk --no-cache add tzdata
WORKDIR $GOPATH/src/wimaha/home-charge/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
#RUN CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 go build -o app /go/bin/home-charge
RUN CGO_ENABLED=1 GOOS=linux go build -o /go/bin/home-charge -a -ldflags '-linkmode external -extldflags "-static"' .
#RUN go build -o /go/bin/home-charge
############################
# STEP 2 build a small image
############################
FROM scratch
# Timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Europe/Berlin
# Copy our static executable.
COPY --from=builder /go/bin/home-charge /home-charge
COPY ./html ./html
COPY ./static ./static
COPY ./settings ./settings
COPY ./database ./database
EXPOSE 7618
# Run the home-charge binary.
ENTRYPOINT ["/home-charge"]