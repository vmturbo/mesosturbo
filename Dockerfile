# Set the base image

FROM ubuntu


# Set the file maintainer

MAINTAINER Pamela Sanchez <pamela.sanchez@turbonomic.com>


ADD mesosturbo /bin/mesosturbo


ENTRYPOINT ["/bin/mesosturbo"]
