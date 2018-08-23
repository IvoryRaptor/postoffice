FROM libratone/librdkafka

LABEL MAINTAINER="libratone"
WORKDIR /opt/docker/project

RUN mkdir -p /opt/docker/project/config.d
#COPY wexin.yaml /opt/docker/project
COPY filetest /opt/docker/project
COPY docker.yaml /opt/docker/project/config.d/

#VOLUME ["/opt/docker/logs"]
EXPOSE 9000
ENTRYPOINT ["/opt/docker/project/filetest"]
CMD []