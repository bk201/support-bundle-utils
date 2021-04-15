FROM nginx:stable

RUN apt-get update && apt-get install -y zip
    
ADD files/nginx/default.conf /etc/nginx/conf.d/default.conf

ADD bin/harvester-sb-collector.sh /usr/bin
RUN chmod +x /usr/bin/harvester-sb-collector.sh

ADD bin/support-bundle-utils /usr/bin
RUN chmod +x /usr/bin/support-bundle-utils
