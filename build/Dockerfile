FROM alpine:3

ARG UCRT_BIN=rtintg-denonavr

LABEL maintainer "Sebastian Plattner"
LABEL org.opencontainers.image.title "Remote Two Integraton for Denon AVR"
LABEL org.opencontainers.image.description "A Unfolded Circle Remote Two integration driver for a Denon AV Receiver"
LABEL org.opencontainers.image.authors "Sebastian Plattner"
LABEL org.opencontainers.image.source https://github.com/splattner/remotetwo-integration-denonavr
LABEL org.opencontainers.image.licenses MPL-2.0

RUN apk --no-cache add libc6-compat

COPY ./$UCRT_BIN /app/rtintg-denonavr

WORKDIR /app

RUN adduser -D nonroot && mkdir /app/ucconfig && chown nonroot /app/ucconfig
USER nonroot

ENTRYPOINT [ "/app/rtintg-denonavr" ]
CMD [ "" ]
