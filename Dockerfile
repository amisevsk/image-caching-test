FROM gcr.io/distroless/base
COPY "./bin/image-caching-test" "/"
CMD ["/image-caching-test"]