FROM golang:1.14.2-buster

RUN GO111MODULE=on go get github.com/google/ko/cmd/ko@v0.5.1
RUN apt update && apt install -y apt-transport-https ca-certificates curl software-properties-common &&\
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - &&\
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable" &&\
    apt update &&\
    apt-cache policy docker-ce &&\
    curl -O https://download.docker.com/linux/ubuntu/dists/bionic/pool/edge/amd64/containerd.io_1.2.2-3_amd64.deb &&\
    apt install ./containerd.io_1.2.2-3_amd64.deb &&\
    apt install -y docker-ce
RUN docker login -u subhobrata.dey@sap.com -p bt74qMgu3eGCjJHQ harbor.eurekacloud.io

VOLUME ["/go/src/kraken.dev/kraken-scheduler"]
WORKDIR /go/src/kraken.dev/kraken-scheduler
COPY . /go/src/kraken.dev/kraken-scheduler

ENV KO_DOCKER_REPO=harbor.eurekacloud.io/eureka
RUN ko publish kraken.dev/kraken-scheduler/cmd/kraken-scheduler-controller
