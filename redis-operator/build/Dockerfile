FROM registry.access.redhat.com/ubi7/ubi-minimal:latest@sha256:244564fff3841542a17abe5d7daf5c803676dc13af48ea5302218b31ee1c2db7

ENV OPERATOR=/usr/local/bin/dbaas-redis-operator \
    USER_UID=1001 \
    USER_NAME=dbaas-redis-operator

# install operator binary
COPY build/_output/bin/dbaas-redis-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
