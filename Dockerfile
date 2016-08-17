# Set the base image

FROM ubuntu


# Set the file maintainer

MAINTAINER Enlin Xu <enlin.xu@turbonomic.com>


ADD mesosturbo /bin/mesosturbo


ENTRYPOINT ["/bin/mesosturbo"]
