CGO_ENABLED=0 go build --installsuffix cgo -a -o mysql-schema-updater . \
&& docker build --tag morfien101/mysql-schema-updater:latest . \
&& docker build --tag morfien101/mysql-schema-updater:$(./mysql-schema-updater -v) . \
&& docker push morfien101/mysql-schema-updater:latest \
&& docker push morfien101/mysql-schema-updater:$(./mysql-schema-updater -v) \
&& rm -f ./mysql-schema-updater