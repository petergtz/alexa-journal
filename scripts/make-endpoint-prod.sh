#!/bin/bash -ex

function_name=AlexaJournal

cd $(dirname $0)/..

for region in 'us-east-1' 'eu-west-1' 'ap-northeast-1'; do
    sed -E -i "s/(arn:aws:lambda:$region:512841817041:function:$function_name)(.*)/\1:prod"'"'"/g" skill-package/skill.json
done

while true; do
    read -p "Deploy changes? " yn
    case $yn in
        [Yy]* ) break;;
        [Nn]* ) exit 1;;
        * ) echo "Please answer yes or no.";;
    esac
done

ask deploy

rm -rf hooks
