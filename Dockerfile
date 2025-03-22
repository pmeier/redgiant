FROM golang:1.22-bookworm AS build

WORKDIR /src

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/redgiant .

FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/redgiant /

ENV REDGIANT_HOST=0.0.0.0
ENV REDGIANT_PORT=80

ENTRYPOINT ["/redgiant"]
CMD ["server"]
# HEALTHCHECK CMD ["/redgiant", "health"]