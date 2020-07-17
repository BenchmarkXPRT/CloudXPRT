FROM ubuntu:18.04

RUN apt-get update
RUN apt-get -y install openssl
RUN apt-get update
RUN apt-get -y upgrade openssl

ADD files /root/app/files

WORKDIR /root/app
ADD cryptserver  /root/app

EXPOSE 8076

CMD ["./cryptserver"]
