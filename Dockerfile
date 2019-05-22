FROM centos

RUN set -x \
  && yum -y update \
  && yum -y install epel-release \
  && yum -y install --enablerepo=epel golang rrdtool-devel

ENV GOPATH /go
WORKDIR /go/src/github.com/YOwatari/grafana-gf-server
COPY main.go .
RUN set -x \
  && go get -v -d ./... \
  && go build
