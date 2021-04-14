FROM nginx:stable

RUN apt-get update && apt-get install -y zip
    
ADD bin/harvester-sb-collector.sh /usr/bin
RUN chmod +x /usr/bin/harvester-sb-collector.sh

ADD files/nginx/default.conf /etc/nginx/conf.d/default.conf
