#!/bin/bash -ex

function_name=AlexaJournal
skill_id=amzn1.ask.skill.ad1669b4-291c-4daa-9fbb-fa32b8ea3078

cd $(dirname $0)/..

for region in 'us-east-1' 'eu-west-1' 'ap-northeast-1'; do
    output=$(aws --region $region lambda publish-version --function-name $function_name)
    echo $output
    version=$(echo $output | jq -r .Version)
    
    aws --region $region lambda update-alias \
      --function-name $function_name \
      --name prod \
      --function-version $version

    aws --region $region lambda add-permission \
      --function-name $function_name:prod \
      --action lambda:invokeFunction \
      --principal alexa-appkit.amazon.com  \
      --statement-id $(date +%s) \
      --event-source-token $skill_id
done
