#!/bin/bash -ex

zip_file=alexa-journal.zip
function_name=AlexaJournal

cd $(dirname $0)/..

pushd cmd/skill
    go build -o main

    rm -f $zip_file
    zip $zip_file main

    aws s3 cp $zip_file s3://alexa-golang-skills/$zip_file
    aws s3 cp s3://alexa-golang-skills/$zip_file s3://alexa-golang-skills-eu-west-1/$zip_file &
    aws s3 cp s3://alexa-golang-skills/$zip_file s3://alexa-golang-skills-ap-northeast-1/$zip_file &
    wait
    rm -f $zip_file
    rm -f main
popd

aws --region us-east-1 lambda update-function-code --function-name $function_name --s3-bucket alexa-golang-skills --s3-key $zip_file &
aws --region eu-west-1 lambda update-function-code --function-name $function_name --s3-bucket alexa-golang-skills-eu-west-1 --s3-key $zip_file &
aws --region ap-northeast-1 lambda update-function-code --function-name $function_name --s3-bucket alexa-golang-skills-ap-northeast-1 --s3-key $zip_file &
wait

