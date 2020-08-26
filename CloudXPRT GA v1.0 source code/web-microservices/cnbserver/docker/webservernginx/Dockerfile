FROM ubuntu:16.04

RUN apt-get update && apt-get install -y --no-install-recommends nginx
ADD nginx.conf /etc/nginx/nginx.conf
ADD cnbserver.crt /etc/ssl/certs/cnbserver.crt
ADD cnbserver.key /etc/ssl/private/cnbserver.key
 
ADD frontend /usr/share/nginx/html

WORKDIR /root/app
ADD cnbserver*  /root/app/
ADD start_web.sh /root/app/

EXPOSE 80 443

CMD ["./start_web.sh"]
