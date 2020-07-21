# Uguisu

![icon](https://user-images.githubusercontent.com/605953/74091901-6d0eef80-4b00-11ea-88c4-b4ae90cd3331.png)

`uguisu` is AWS CDK Construct to monitor suspicious activity regarding AWS resource. `uguisu` watches CloudTrail logs and monitors changes of AWS resources. It also have rules to detect an event of interest regarding security. A part of rules is based on AWS CIS benchmark. `uguisu` notifies detail to Slack channel when detecting an event of interest.

By the way, the name of the tool comes from *uguisubari (鶯張り)* that is floors to alarm someone is incoming by a chirping sound when walked upon. In English, it is called *Nightingale floor*. See [wikipedia](https://en.wikipedia.org/wiki/Nightingale_floor) for more detail.

# How to use

## 0. Prerequisites

### CDK tools

See official getting started page. https://docs.aws.amazon.com/cdk/latest/guide/getting_started.html. Please install CDK tools.

### Slack Incoming Webhook URL

See https://api.slack.com/messaging/webhooks to create your Incoming Webhook URL. You can get URL like this:

```
https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
```

### Setup CloudTrail logging to S3 and SNS topic

Also CloudTrail logs are required to monitor AWS resources. `uguisu` requires not only CloudTrail logs but also SNS topic to notify `s3:ObjectCreated:*` event from S3 bucket.

- Enable CloudTrail and logging to S3: https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-create-a-trail-using-the-console-first-time.html
- Change S3 bucket policy: https://docs.aws.amazon.com/awscloudtrail/latest/userguide/create-s3-bucket-policy-for-cloudtrail.html
- Configure S3 event notification to SNS: https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html

## 1. Create your new CDK project

```bash
$ mkdir your-cdk-app
$ cd your-cdk-app
$ cdk init --language typescript
```

## 2. Install Uguisu module

```bash
# install from github
$ npm install m-mizutani/uguisu
```

## 3. Write your construct

Put construct code to `bin/your-cdk-app.ts` like following. Please replace `snsTopicARN`, `lambdaRoleARN` and `slackWebhookURL`.

```ts
#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { UguisuStack } from "uguisu";

const app = new cdk.App();
new UguisuStack(app, "secops-uguisu", {
  snsTopicARN: "arn:aws:sns:ap-northeast-1:1234567890:your-cloudtrail-event-topic",
  lambdaRoleARN: "arn:aws:iam::1234567890:role/YourLambdaRole",
  slackWebhookURL:
    "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
});
```

## 4. Deploy your construct

```bash
$ npm run build
$ cdk deploy
```

# License

MIT License