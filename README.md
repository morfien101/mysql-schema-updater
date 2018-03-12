# mysql-schema-updater

## What is it?

A tool to update MySQL schema in projects.

This tool can be used to bootstrap MySQL databases for your projects. It can be used inside a docker container or as a standalone binary. The application will read either os commands or environment variables.

The container can be pulled from [DockerHub | morfien101/mysql-schema-updater](https://hub.docker.com/r/morfien101/mysql-schema-updater/)

## How to use it

The below table will explain the environment variables and how they map to a command line:

Flag | Env var | Default | Description
---|---|---|---
sqlhost | SQL_HOST | localhost | Hostname or IP  of the SQL service you want to talk to.
sqlport | SQL_PORT | 3306 | Port number used while talking to the server.
sqlusername | SQL_USERNAME | root | Username to use when connecting to the SQL service.
sqlpassword | SQL_PASSWORD | root | Password to use when connecting to the SQL service.
sqldb | SQL_DB | test | Name of the database that should be used. If used in conjuntion with create-db it will make a DB with this name.
sqlversion | SQL_VERSION_TABLE | version | The name of the version table to use. If used with create-db it will create this table also.
scripts-path | SCRIPTS_PATH | /data | Where the sql files are kept.
create-db | CREATE_DB | false | Should the tool try to bootstrap the server for you also.
use-environment-variables | - | Use this flag if you want the application to read environment variables. Used primarly with docker containers on orchastrated platforms.
v | - | Shows the version of the application.
h | - | Shows the help menu.

Its its naked form the tool can be used with os arguments. See below for an example.

```bash
./mysql-schema-updater \
  --create-db \
  --scripts-path ./testsqlfiles \
  --sqldb go_is_awesome \
  --sqlhost 127.0.0.1 \
  --sqlpassword P4ssword \
  --sqlusername root
```

In a docker container you can use the following syntax:

```bash
# Whats in the environments file
cat fullEnvList
# Will produce:
# SQL_HOST=172.17.0.1
# SQL_PORT=3306
# SQL_USERNAME=root
# SQL_PASSWORD=P4ssword
# SQL_DB=go_is_awesome
# CREATE_DB=true

# Run the docker command in terminal mode to see the output.
docker run -it \
  --env-file $(pwd)/test_scripts/fullEnvList \
  -v $(pwd)/testsqlfiles:/data \
  morfien101/mysql-schema-updater:latest
```

The container used is started from __scratch__ so there is noting in there. You will not be able to use bash to see inside the container. Please also be aware of the mount to /data. This is the default location of where to look for your .sql files.

ALL FILES MUST END IN .sql TO BE PICKED UP.

All the sql statements will be printed to the terminal so that you can debug any crashes.
Important if you have sensitve data in there.
It is not currently supported to supress this printing.

## Why do we need it

This tool is primarly used in a situation where you want your updates to be automated. It will read files in a folder and process them. It will also check to see what versions are needed and only apply those files. You can mix this with something that will pull files from something like github and process the ones that are required.