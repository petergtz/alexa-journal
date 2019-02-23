#!/bin/bash -e

. private/dev-application-id.sh
. private/github-token.sh

export cert=private/certificate.pem
export key=private/private-key.pem
fresh
