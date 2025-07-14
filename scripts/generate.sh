SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
pushd "${SCRIPT_DIR}" > /dev/null 2>&1 || exit

printf "Generating metrics scrape file for test fixture... "
pushd ../tools/generate-fixture > /dev/null 2>&1 || exit
go run . --input ../../data/master.csv > ../../integrations-extras/redpanda/tests/fixtures/redpanda_metrics.txt
popd > /dev/null 2>&1 || exit
printf "done.\n"

printf "Generating metadata.csv... "
pushd ../tools/generate-metadata > /dev/null 2>&1 || exit
go run . --input ../../data/master.csv > ../../integrations-extras/redpanda/metadata.csv
popd > /dev/null 2>&1 || exit
printf "done.\n"

popd > /dev/null 2>&1 || exit