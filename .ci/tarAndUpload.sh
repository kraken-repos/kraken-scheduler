#!/usr/bin/env bash
export MEDIA_TOKEN='MSxrcmFrZW4tc2NoZWR1bGVyL2V1cmVrYSxaRWRXZEdOSGVHaGtSMVYwV2xjMWJtRlhOV3c9LDE1OTY0Nzc1NTgsNDg0NDYyNjE1OTE5NDIwNzEyLDYyNzA4Y2RkMGU1NWM0NjliODFkMTMxZDE1MGYyOWRl'
#set -x

lower() {
    # Usage: lower "string"
    printf '%s\n' "${1,,}"
}

collect_files () {
    wk_dir=$1
    rpt_dir=$2
    dir_pat="$3-*"
    col_dir=$4

    [ -e $rpt_dir ] && rm -rf $wk_dir/$rpt_dir
    mkdir -p $wk_dir/$rpt_dir

    for dir in $wk_dir/$dir_pat/; do
        [ -e $wk_dir/$dir/$col_dir ] && cp -r $wk_dir/$dir/$col_dir $rpt_dir/$dir
    done
}

tar_upload () {
    wk_dir=$1
    upload_dir=$2
    sub_url=$3
    result_name=$4

    tar_name=upload.tar.gz

    tar -czf $wk_dir/$tar_name -C $wk_dir/$upload_dir .
    curl --fail -X POST -H "Media-Token: $MEDIA_TOKEN" -H  "md5:`md5sum -b $wk_dir/$tar_name | sed 's/ .*//'`" --data-binary @$wk_dir/$tar_name \
         "https://media.eurekacloud.io/artifacts/$sub_url/"


    echo "$result_name result url:"
    echo "https://media.eurekacloud.io/static/$sub_url/"
}

report_to_pr () {
    result_name=$1
    sub_url=$2

    visit_url="https://media.eurekacloud.io/static/$sub_url/"
    result_name_l=`lower "$result_name"`

    echo "Send result link to collect Server:"

    # '{"types":[{"msg_type":"Unit Test Results","content":[{"msg_key":"Unit test result","msg_value":"'$visit_url'"}]}]}'

    p_1='{"types":[{"msg_type":"'
    p_2=' Results","content":[{"msg_key":"'
    p_3=' result","msg_value":"'
    p_4='"}]}]}'

    curl -d "$p_1$result_name$p_2$result_name_l$p_3$visit_url$p_4" \
         -H "Content-Type: application/json" -X POST http://pr-comments/collect

}

if [ $# -eq 4 ]; then
    g_rpt_dir='report'

    g_repo_name=$1
    g_sub_dir=$2
    g_sub_url=$3
    g_result_name=$4

    collect_files '.' $g_rpt_dir $g_repo_name "$g_sub_dir"
    tar_upload '.' $g_rpt_dir "$g_repo_name/$g_sub_url" "$g_result_name"
    report_to_pr "$g_result_name" "$g_repo_name/$g_sub_url"
else
    echo "call with repo_name report_dir \$GIT_COMMIT/unit_test \"unit test\""
fi
