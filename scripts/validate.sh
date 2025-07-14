SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
pushd "${SCRIPT_DIR}" > /dev/null 2>&1 || exit 1

printf "Validating metrics.py... "
pushd ../tools/validate-metrics-py > /dev/null 2>&1 || exit 1
go run . --master ../../data/master.csv --code ../../integrations-extras/redpanda/datadog_checks/redpanda/metrics.py || exit 1
popd > /dev/null 2>&1 || exit 1
printf "done.\n"

echo "Validating common.py... "
pushd ../tools/validate-common-py > /dev/null 2>&1 || exit 1
go run . --master ../../data/master.csv --code ../../integrations-extras/redpanda/tests/common.py || exit 1
popd > /dev/null 2>&1 || exit 1
printf "done.\n"

printf "Validating rp metrics in master.csv against public docs... "
pushd ../tools/check-coverage > /dev/null 2>&1 || exit 1
go run . --master ../../data/master.csv  || exit 1
popd > /dev/null 2>&1 || exit 1
printf "done.\n"

popd > /dev/null 2>&1 || exit 1