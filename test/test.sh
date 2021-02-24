###
 # @Author: JiHan
 # @Date: 2021-02-23 17:18:39
 # @LastEditTime: 2021-02-24 11:05:30
 # @LastEditors: JiHan
 # @Description: Test goipset client
 # @Usage: 
### 

set -e

#destroy all ipset list
ipset list | grep Name | awk -F': ' '{print $2}' | xargs -t -i ipset destroy {}

if [ ! -f ./goipset ];then
    go build -o goipset main.go
fi

while read line
do
    if [ "${line}" == "" ]; then
        continue
    fi
    echo "========Run:${line}=========="
    ./goipset ${line}
    echo "------------------"
    ipset list
done < ./test.date