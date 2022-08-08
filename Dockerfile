FROM golang:1.17 as builder
ARG BEAT_VERSION=v7.17.5
RUN apt-get update && apt-get install patch -y
RUN git clone -b $BEAT_VERSION https://github.com/elastic/beats.git
WORKDIR beats
COPY metricbeat_patch.diff .
RUN patch -p1 < metricbeat_patch.diff
WORKDIR metricbeat
ADD zebrium module/zebrium
RUN ls module/
RUN ls module/zebrium
RUN rm -rf modules.d
ADD zebrium/_meta/zebrium.yml modules.d/
RUN go install


FROM golang:1.17
ARG user=zebeat
ARG group=zebeat
ARG uid=1000
ARG gid=1000
ARG ZEBEAT_HOME=/var/zebeat
ENV ZEBEAT_HOME $ZEBEAT_HOME
ENV PATH "${ZEBEAT_HOME}:${PATH}"

RUN mkdir -p $ZEBEAT_HOME \
  && chown ${uid}:${gid} $ZEBEAT_HOME \
  && groupadd -g ${gid} ${group} \
  && useradd -N -d "$ZEBEAT_HOME" -u ${uid} -g ${gid} -m -s /bin/bash ${user}
USER ${user}
WORKDIR $ZEBEAT_HOME
COPY --from=builder go/bin/metricbeat .
ADD  docker/docker-entrypoint.sh .
ENTRYPOINT ["docker-entrypoint.sh"]