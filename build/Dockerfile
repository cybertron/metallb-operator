FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/metallb-operator \
    USER_UID=1001 \
    USER_NAME=metallb-operator

# install operator binary
COPY build/_output/bin/metallb-operator ${OPERATOR}
# install metallb manifests
COPY manifests/namespace.yaml /manifests/namespace/
COPY manifests/metallb.yaml /manifests/metallb/
COPY manifests/config.yaml /manifests/config/

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
