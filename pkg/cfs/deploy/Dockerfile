FROM centos

RUN yum install -y fuse

ADD cfs-fuseclient /usr/bin/
ADD cfsplugin / 

ENTRYPOINT ["/cfsplugin"]
