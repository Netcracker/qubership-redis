FROM ghcr.io/netcracker/qubership-docker-integration-tests:0.5.6

ENV ROBOT_OUTPUT=${ROBOT_HOME}/output \
    DISTR_DIR=/tmp/deps

RUN mkdir -p ${ROBOT_HOME} \
    && mkdir -p ${ROBOT_OUTPUT}

COPY requirements.txt ${ROBOT_HOME}/requirements.txt
COPY robot ${ROBOT_HOME}
COPY entrypoint.sh /entrypoint.sh

RUN apk update \
    && apk add --no-cache build-base \
    && apk upgrade --no-cache \
    && rm -rf /var/cache/apk/*

# Upgrade all tools to avoid vulnerabilities
RUN set -x && apk upgrade --no-cache --available

RUN set -x \
    && pip3 install --upgrade pip setuptools \
    && pip3 install -r ${ROBOT_HOME}/requirements.txt \
    && rm -rf /var/cache/apk/*


RUN chmod -R 777 ${ROBOT_HOME} \
    && chmod +x /entrypoint.sh

ENV PYTHONDONTWRITEBYTECODE=1

USER 1000

# Expose the port
EXPOSE 8080
VOLUME ["${ROBOT_HOME}"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["run-robot"]
