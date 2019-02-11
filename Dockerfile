FROM gcr.io/distroless/base
COPY "./bin/che-image-caching" "/"
CMD ["/che-image-caching"]
