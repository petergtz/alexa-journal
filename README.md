# Alexa Journal

## Development

### Running test suite

#### Getting `GOOGLE_DRIVE_TOKEN`

Go to https://developer.amazon.com/alexa/console/ask/test/amzn1.ask.skill.ad1669b4-291c-4daa-9fbb-fa32b8ea3078/development/de_DE/, make a sample request, and look for `accessToken` in JSON input.

Copy it and paste it into `token.sh`.

Then:
```
source private/token.sh
```

To run tests:
```
ginkgo -r
```

### Making changes and publish new code

#### Changes in the Lambda Service

New code gets pushed via `scripts/deploy-code.sh`. That doesn't make the code available in Prod, i.e. it's not customer facing, but it can be tested in the Alexa developer console, given that the endpoint there is pointing to the non-prod endpoint. To publish the deployed code in prod, there is `publish-version`. That will create a new version in Lambda and also move the prod alias to that latest version.

Workflow:
1. Make changes
1. Deploy changes `scripts/deploy-code.sh`
3. Verify changes work in the [console](https://developer.amazon.com/alexa/console/ask/test/amzn1.ask.skill.ad1669b4-291c-4daa-9fbb-fa32b8ea3078/development/de_DE/)
4. When everything looks good, deploy to production: `scripts/publish-version.sh`

#### Changes in the Alexa Model