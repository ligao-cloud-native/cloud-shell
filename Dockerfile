FROM debian:stable-slim

ADD kubectl/v1.19.2/kubexx/kubectl /usr/local/bin/
RUN chmod 755 /usr/local/bin/kubectl
WORKDIR /root