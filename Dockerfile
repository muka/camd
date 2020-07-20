FROM scratch
ARG ARCH=amd64
ADD ./build/camd-${ARCH} /camd
ENTRYPOINT ["/camd"]
