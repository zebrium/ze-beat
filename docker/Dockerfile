ARG RELEASE=latest
FROM 872295030327.dkr.ecr.us-west-2.amazonaws.com/zebrium/baseos:$RELEASE

# TODO: remove sudo user before deployment on AWS
RUN echo 'sudo:x:27:zebrium' >> /etc/group; echo 'zebrium:Zebrium#1' | chpasswd

COPY --chown=1000:1000 ./bashrc /home/zebrium/.bashrc
RUN cat /home/zebrium/.bashrc >> /root/.bashrc

COPY --chown=1000:1000 ./metricbeat /usr/local/bin
COPY --chown=1000:1000 ./docker-entrypoint.sh /usr/local/bin

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
