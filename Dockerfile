# This file must be used after the "build.sh" compilation is performed.

# Set up base image and set meta information.
FROM alpine:latest
MAINTAINER fooage<nullptr.wang@gmail.com>
ENV version 1.0

# Copy the compiled product to the image.
WORKDIR /app
COPY output/ /app/

# Change the system time zone of the mirror to Shanghai and set UTF-8 encoding.
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone
ENV LANG C.UTF-8

# Start the service after mapping the port.
EXPOSE 8090
ENTRYPOINT ["/app/benzene"]