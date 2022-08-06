# -----------------------------------------------------------------------------
# Usage: `docker run -it --privileged <IMAGE>`. Make sure to add `-t` and `--privileged`.

ARG GO_VERSION=1.17

FROM golang:${GO_VERSION}-bullseye

RUN apt update && \
    apt install -y iproute2

# setup netns and veth pair
RUN ip netns add netns0 && \
    ip link add veth0 type veth peer name ceth0 && \
    ip link set veth0 up && \
    ip link set ceth0 netns netns0 && \
    ip exec netns0 ip link set lo up && \
    ip exec netns0 ip link set ceth0 up && \
    ip netns add netns1 && \
    ip link add veth1 type veth peer name ceth1 && \
    ip link set veth1 up && \
    ip link set ceth1 netns netns1 && \
    ip exec netns1  ip link set lo up && \
    ip exec netns1  ip link set ceth1 up && \
    ip link add br0 type bridge && \
    ip link set br0 up && \
    ip link set veth0 master br0 && \
    ip link set veth1 master br0

COPY . /go/src/github.com/t1anz0ng/iftree
WORKDIR /go/src/github.com/t1anz0ng/iftree
RUN go build -o iftree && \
    install -D -m 755 $(CURDIR)/iftree /out/bin/iftree




