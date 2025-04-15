buf-generate:
    buf dep update && \
    buf generate

buf-push-test:
    buf dep update && \
    buf build && \
    buf push --label v1 --label testrelease