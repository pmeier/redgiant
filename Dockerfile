ARG PYTHON_TAG=3.10-slim
FROM python:$PYTHON_TAG AS python

FROM python AS builder

WORKDIR /red-giant

RUN pip install --progress-bar=off virtualenv && \
    virtualenv venv && \
    ./venv/bin/python -m pip install build
ENV PATH="/red-giant/venv/bin:${PATH}"

COPY requirements.lock .
RUN pip install --progress-bar=off --no-deps --no-cache \
    --requirement requirements.lock

COPY . ./
ARG VERSION
RUN SETUPTOOLS_SCM_PRETEND_VERSION=$VERSION \
    pip install --progress-bar=off --no-deps --no-cache \
    .

FROM python

RUN useradd --no-create-home --no-log-init --shell $(which bash) solarian
USER solarian
WORKDIR /red-giant

COPY --from=builder --chown=solarian:solarian /red-giant/venv /red-giant/venv
ENV PATH="/red-giant/venv/bin:${PATH}"

ENV RED_GIANT_HOST=0.0.0.0
ENV RED_GIANT_PORT=80
EXPOSE ${RED_GIANT_PORT}/tcp

ENTRYPOINT ["red-giant"]
CMD ["serve"]
HEALTHCHECK CMD ["red-giant", "healthcheck"]
