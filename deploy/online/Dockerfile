FROM debian:buster-slim

RUN set -eux; \
    sed -i s@/deb.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list ; \
    sed -i s@/security.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo 'Asia/Shanghai' >/etc/timezone

# persistent / runtime deps
RUN set -eux; \
	apt-get update; \
	apt-get install -y --no-install-recommends \
		ca-certificates \
		curl \
		telnet \
        iputils-ping \
        net-tools \
		iproute2 \
	; \
	rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    mkdir -p /eventmesh/config/templates; \
    mkdir -p /home/logs/xeslog && chmod 755 /home/logs/xeslog ; \
    mkdir -p /data/eventmesh; \
    ln -sf /home/logs/xeslog /data/eventmesh/logs

ADD ./config.yml /eventmesh/config/config.yml
ADD ./route.yml /eventmesh/config
ADD ./default.tmpl /eventmesh/config/templates
ADD ./eventroute /eventmesh/

RUN chmod u+x  /eventmesh/eventroute

WORKDIR /eventmesh/
EXPOSE 8080
CMD ["/eventmesh/eventroute"]
