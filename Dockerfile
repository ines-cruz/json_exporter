

FROM golang:1.15 AS build

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apt install git
RUN go get github.com/ines-cruz/json_exporter

# List project dependencies with Gopkg.toml and Gopkg.lock
# These layers are only re-built when Gopkg files are updated
COPY Gopkg.lock Gopkg.toml /go/src/json_exporter/
WORKDIR /go/src/json_exporter/
# Install library dependencies
RUN dep ensure -vendor-only

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o /bin/json_exporter

# This results in a single layer image
FROM scratch
COPY --from=build /bin/json_exporter /bin/json_exporter
ENTRYPOINT ["/bin/json_exporter"]
CMD ["--help"]
