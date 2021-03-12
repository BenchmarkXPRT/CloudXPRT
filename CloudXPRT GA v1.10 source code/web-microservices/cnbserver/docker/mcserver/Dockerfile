FROM ubuntu:18.04

ADD plugin /root/app/plugin
ENV PATH="/root/app/plugin:${PATH}"
ENV LD_LIBRARY_PATH=/root/app/plugin
ENV OMP_NUM_THREADS=4

WORKDIR /root/app
ADD mcserver  /root/app
ADD MonteCarlo*  /root/app/

EXPOSE 8074

CMD ["./mcserver"]
