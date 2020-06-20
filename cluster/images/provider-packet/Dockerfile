FROM BASEIMAGE
RUN apk --no-cache add ca-certificates bash

ARG ARCH
ARG TINI_VERSION

ADD package /usr/local/bin/crossplane-packet-package

COPY package /

EXPOSE 8080
ENTRYPOINT ["crossplane-packet-provider"]
