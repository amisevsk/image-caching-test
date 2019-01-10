FROM gcr.io/distroless/base
COPY "./image-caching-test" "/"
CMD ["/image-caching-test"]