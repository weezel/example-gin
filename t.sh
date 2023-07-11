#!/usr/bin/env bash

# Integration tests with cover profiled binary.
# More info here https://go.dev/blog/integration-test-coverage

# Bail out on errors, important in integration testing
set -euo pipefail

readonly app_dir="cmd/webserver"
readonly bin_name="example-gin-cover"
readonly cov_data_dir="covdatafiles"

pkill -f "${bin_name}" || true
rm -rf "${bin_name}" "${cov_data_dir}" cov.txt || true
mkdir "${cov_data_dir}"
go build -race -cover -o "${bin_name}" "${app_dir}/main.go"
GOCOVERDIR="${cov_data_dir}" ./"${bin_name}" &

# Wait a bit until the web server is launched
sleep 2

# This is a simple demo scripts showing how to create, list and delete users.
printf "### Add users Sebastian, Laura and Cecilia\n"
curl -d '{"name": "Sebastian", "age": 24}' -X POST "localhost:8080/user"
curl -d '{"name": "Laura", "age": 26}' -X POST "localhost:8080/user"
curl -d '{"name": "Cecilia", "age": 27}' -X POST "localhost:8080/user"

printf "\n### Get users\n"
#curl -X GET "localhost:8080/user/"
curl -X GET "localhost:8080/user/Sebastian"
curl -X GET "localhost:8080/user/Laura"
curl -X GET "localhost:8080/user/Cecilia"

printf "\n### Delete users\n"
curl -d '{"name": "Sebastian"}' -X DELETE "localhost:8080/user"
echo
curl -d '{"name": "Laura"}' -X DELETE "localhost:8080/user"
echo
curl -d '{"name": "Cecilia"}' -X DELETE "localhost:8080/user"
echo

printf "\n### We should have zero users now:\n"
curl -X GET "localhost:8080/user/"
echo

pkill -f "${bin_name}"

# Post-process the coverage data files and convert to txt file
go tool covdata percent -i="${cov_data_dir}"
go tool covdata textfmt -i="${cov_data_dir}" -o=cov.txt
go tool cover -func=cov.txt

# Merge different runs of coverage results
rm -rf merged_cov_data
mkdir merged_cov_data
go tool covdata merge -i="${cov_data_dir}" -o=merged_cov_data

