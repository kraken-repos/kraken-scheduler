#!/usr/bin/env bash
export MEDIA_TOKEN='MSxrcmFrZW4tc2NoZWR1bGVyL2V1cmVrYSxaRWRXZEdOSGVHaGtSMVYwV2xjMWJtRlhOV3c9LDE1OTY0Nzc1NTgsNDg0NDYyNjE1OTE5NDIwNzEyLDYyNzA4Y2RkMGU1NWM0NjliODFkMTMxZDE1MGYyOWRl'
#set -ex

rpt_dir="report"

[ -e ./$rpt_dir ] || mkdir ./$rpt_dir

for dir in ./kraken-scheduler-*/; do
    if [ -e ./$dir/build/reports/tests/test ]; then
        mkdir -p ./$rpt_dir/$dir/
        cp -r ./$dir/build/reports/tests/test ./$rpt_dir/$dir/
    fi
done

tar -czf ./$rpt_dir.tar.gz -C ./$rpt_dir .

sub_utl=kraken-scheduler/$GIT_COMMIT/UnitTest

url="https://media.eurekacloud.io/artifacts/$sub_utl/"
curl --fail -X POST -H "Media-Token: $MEDIA_TOKEN" -H  "md5:`md5sum -b ./$rpt_dir.tar.gz | sed 's/ .*//'`" --data-binary @./$rpt_dir.tar.gz \
     $url


echo "unit test result url:"
visit_url="https://media.eurekacloud.io/static/$sub_utl/"
echo $visit_url

echo "Send result link to collect Server:"
curl -d '{"types":[{"msg_type":"Unit Test Results","content":[{"msg_key":"Unit test result","msg_value":"'$visit_url'"}]}]}' \
     -H "Content-Type: application/json" -X POST http://pr-comments/collect

