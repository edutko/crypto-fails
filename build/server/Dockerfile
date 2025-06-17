FROM --platform=$BUILDPLATFORM golang:1.24 AS build
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=0

WORKDIR /src
RUN --mount=type=bind,source=.,target=/src \
    make -e OUT=/out /out/server \
    && cp -r /src/web /out/

FROM scratch
ENV LISTEN_ADDR=":8080"
COPY --from=build /out/server /
COPY --from=build /out/web /web
VOLUME /data
WORKDIR /
ENTRYPOINT ["/server"]
