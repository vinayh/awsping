FROM golang:1.26-bookworm AS build
COPY . /build
WORKDIR /build
RUN make

FROM gcr.io/distroless/base
COPY --from=build /build/awsping /

ENTRYPOINT ["/awsping"]
