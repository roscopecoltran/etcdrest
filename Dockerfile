FROM gliderlabs/alpine

COPY bin/sampleapp-docker /sampleapp-docker
COPY src/github.com/docker/sampleapp-docker/templates /templates

EXPOSE 3000
CMD ["/sampleapp-docker"]
