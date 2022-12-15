FROM envoyproxy/envoy-dev:latest

RUN apt-get update && apt-get -q install -y \
    curl

RUN chmod go+r /etc/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy.yaml