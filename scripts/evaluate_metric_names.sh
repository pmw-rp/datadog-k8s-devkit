SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
pushd "${SCRIPT_DIR}/../tools/check-levenstein" || exit

go run . --master ../../data/master.csv | grep -v '2,' | sort -nr  || exit 1

popd > /dev/null 2>&1 || exit