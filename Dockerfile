FROM ubuntu:xenial

RUN apt-get -y update

RUN apt-get -y install \
  gcc \
  binutils \
  make \
  perl \
  liblzma-dev \
  mtools \
  mkisofs \
  isolinux \
  syslinux \
  git \
;

RUN apt-get -y install \
  linux-image-generic \
;

RUN mkdir /work

ENV IPXE_GITSHA "36a4c85f911c85f5ab183331ff74d125f9a9ed32"

RUN git clone git://git.ipxe.org/ipxe.git /work/ipxe \
    && cd /work/ipxe \
    && git checkout $IPXE_GITSHA

# cache build objects
RUN cd /work/ipxe/src && \
    make
