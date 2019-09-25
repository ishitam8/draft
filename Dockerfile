FROM alpine:latest
RUN apk add --no-cache bash git curl
RUN mkdir -p /draft/plugins
COPY ./_dist/linux-amd64/draft /bin
ENV DRAFT_HOME=/draft
RUN ["draft", "init"]