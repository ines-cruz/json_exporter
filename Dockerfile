

FROM golang:1.15 AS build

# Install tools required for project
# Run `docker build --no-cache .` to update dependencies
RUN apt install git
RUN go get -u github.com/ines-cruz/json_exporter


WORKDIR /go/src/json_exporter/

EXPOSE 7979

# Copy the entire project and build it
# This layer is rebuilt when a file changes in the project directory
COPY . /go/src/json_exporter/
RUN go build -o /bin/json_exporter

# This results in a single layer image
FROM ubuntu:18.04

RUN apt-get update && apt-get -y install curl
COPY --from=build bin/json_exporter /json_exporter


ENTRYPOINT ["/json_exporter"]

CMD ["python -m SimpleHTTPServer 8080 &"]
CMD [" http://localhost:8080/example/output.json example/config.yml & "]
