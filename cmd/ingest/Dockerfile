FROM golang:1.24 AS build

# Set destination for COPY
WORKDIR /

# Download Go modules
COPY ../../go.mod ../../go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY ../../ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ingest

FROM alpine:latest AS release

WORKDIR /

COPY --from=build /ingest /ingest

ENTRYPOINT [ "/ingest" ]