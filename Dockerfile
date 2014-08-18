FROM golang:1.3
ADD . /go/src/github.com/cpuguy83/lestrade
RUN cd /go/src/github.com/cpuguy83/lestrade && go get && go build
ENTRYPOINT ["/go/src/github.com/cpuguy83/lestrade/lestrade"]
CMD ["-g", "/docker"]
