FROM golang:1.23-bookworm AS build

WORKDIR /src

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/redgiant ./main

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/redgiant /

ENV REDGIANT_HOST=0.0.0.0

ENTRYPOINT ["/redgiant"]
CMD ["serve"]
HEALTHCHECK CMD ["/redgiant", "health"]
