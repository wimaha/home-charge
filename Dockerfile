############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/wimaha/home-charge/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN go build -o /go/bin/home-charge
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/home-charge /home-charge
COPY ./html ./html
COPY ./static ./static
COPY ./settings ./settings
EXPOSE 7618
# Run the home-charge binary.
ENTRYPOINT ["/home-charge"]