FROM centos:stream8
MAINTAINER losuler <losuler@posteo.net>

WORKDIR /opt/futaba

RUN dnf -y install golang

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY main.go ./
RUN go build -o futaba main.go

COPY config.yml.example /etc/futaba.yml

CMD ["./futaba"]
