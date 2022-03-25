FROM golang:1.17 as builder
RUN apt-get update && apt-get install patch -y
RUN git clone https://github.com/elastic/beats.git
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
WORKDIR /usr/local/bin
COPY --from=builder go/bin/metricbeat .
ADD  docker/docker-entrypoint.sh .
ENTRYPOINT ["docker-entrypoint.sh"]
