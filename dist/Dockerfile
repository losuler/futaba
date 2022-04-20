FROM centos:stream8
MAINTAINER losuler <losuler@posteo.net>

WORKDIR /opt/futaba

RUN dnf -y install golang git
RUN git clone https://gitlab.com/losuler/futaba .

RUN go mod download
RUN go build -o futaba main.go

COPY config.yml.example /etc/futaba.yml

CMD ["./futaba"]
