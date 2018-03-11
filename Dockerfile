FROM scratch
ADD /dummyfile /data/dummyfile
ADD ./mysql-schema-updater /mysql-schema-updater
ENTRYPOINT ["/mysql-schema-updater"]
CMD ["-use-environment-variables"]