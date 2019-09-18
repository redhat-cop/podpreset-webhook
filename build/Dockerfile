FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/podpreset-webhook \
    USER_UID=1001 \
    USER_NAME=podpreset-webhook

# install operator binary
COPY build/_output/bin/podpreset-webhook ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
