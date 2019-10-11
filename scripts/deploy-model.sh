#!/bin/bash -ex

ask diff  --target model

while true; do
    read -p "Deploy changes? " yn
    case $yn in
        [Yy]* ) break;;
        [Nn]* ) exit 1;;
        * ) echo "Please answer yes or no.";;
    esac
done

ask deploy  --target model

rm -rf hooks