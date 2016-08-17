# Set the base image
FROM ubuntu

# Set the file maintainer
MAINTAINER Pamela Sanchez <pamela.sanchez@vmturbo.com>

COPY ./go_app/mesosturbo /bin/mesosturbo

