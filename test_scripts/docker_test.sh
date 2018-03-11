docker run -it \
  --env-file $(pwd)/test_scripts/fullEnvList \
  -e SCRIPTS_PATH=/data \
  -v $(pwd)/testsqlfiles:/data \
  morfien101/mysql-schema-updater:latest