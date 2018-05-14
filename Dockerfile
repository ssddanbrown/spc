FROM alpine:3.7
LABEL maintainer="ssddanbrown@googlemail.com"
ADD spc spc
RUN apk add --no-cache ca-certificates
CMD ["/spc"]