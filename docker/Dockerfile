FROM gcr.io/datadoghq/agent:latest

COPY datadog_redpanda-*-py2.py3-none-any.whl /tmp
RUN agent integration install -v -r -w /tmp/datadog_redpanda-*-py2.py3-none-any.whl