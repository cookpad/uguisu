# Uguisu

![icon](https://user-images.githubusercontent.com/605953/74091901-6d0eef80-4b00-11ea-88c4-b4ae90cd3331.png)

`uguisu` is an AWS CDK Construct that watches CloudTrail logs and sends Slack notifications when events of interest occur. Rules cover resource lifecycle events, security service tampering, and the AWS CIS Benchmark monitoring controls.

<img width="657" alt="uguisu" src="https://user-images.githubusercontent.com/605953/88273381-147d8880-cd15-11ea-8403-1125f4bed14f.png">


The name comes from *uguisubari (鶯張り)* - floors that make a chirping sound when walked upon, alerting to intruders. In English, this is called a *Nightingale floor*. See [wikipedia](https://en.wikipedia.org/wiki/Nightingale_floor) for more detail.

# Rules

- Based on AWS CIS Benchmark
  - 3.1: Unauthorized API calls monitoring
  - 3.2: Management Console sign-in without MFA
  - 3.3: Usage of root account (write actions only)
  - 3.4: IAM policy changes
  - 3.5: CloudTrail configuration changes
  - 3.6: AWS Management Console authentication failures
  - 3.7: Disabling or scheduled deletion of customer created CMKs
  - 3.8: S3 bucket policy changes
  - 3.9: AWS Config configuration changes
  - 3.10: Security group changes
  - 3.11: Network Access Control Lists (NACL)
  - 3.12: Changes to network gateways
  - 3.13: Route table changes
  - 3.14: VPC changes
- Resource life events
  - ACM: Certificate export, import, renew, or delete
  - EC2: Instance launch or termination (excludes autoscaling, batch, and ECS)
  - EKS: Cluster creation or deletion
  - IAM: User/role creation or deletion, access key changes, login profile changes, group membership changes
  - Lambda: Function creation, deletion, code updates, or permission changes
  - RDS: Instance creation or deletion
  - S3: Bucket creation or deletion
  - Secrets Manager: Secret creation, deletion, updates, rotation, or resource policy changes
  - VPC: New VPC created
  - Organization/Account: Account and organization lifecycle events
- Security service tampering
  - GuardDuty detector deletion or disassociation
  - Security Hub disabled or insights deleted
  - CloudWatch alarms deleted or disabled


# How to use

## 0. Prerequisites

### CDK tools

See official getting started page. https://docs.aws.amazon.com/cdk/latest/guide/getting_started.html. Please install CDK tools. Note that currently, CDK v2 is not supported. Please use CDK v1.

### Slack Incoming Webhook URL

See https://api.slack.com/messaging/webhooks to create your Incoming Webhook URL. You will get a URL like this:

```
https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX
```

### Setup CloudTrail logging to S3 and SNS topic

CloudTrail logs are required to monitor AWS resources. `uguisu` requires not only CloudTrail logs but also an SNS topic to notify `s3:ObjectCreated:*` event from S3 bucket.

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
$ npm install uguisu
```

## 3. Write your construct

Put construct code to `bin/your-cdk-app.ts` like the following. Replace `s3BucketName`, `snsTopicARN`, `lambdaBuildPath`, `lambdaPackagePath`, and `slackWebhookURL` with your values.

```ts
#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { UguisuStack } from "uguisu";

const app = new cdk.App();
new UguisuStack(app, "secops-uguisu", {
  lambdaBuildPath: "./",
  lambdaPackagePath: "./lambda/tracker",
  s3BucketName: "your-cloudtrail-logs-bucket",
  snsTopicARN: "arn:aws:sns:ap-northeast-1:1234567890:your-cloudtrail-event-topic",
  slackWebhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
});
```

### Optional parameters

| Parameter | Description |
|---|---|
| `lambdaRoleARN` | ARN of an existing IAM role to use for the Lambda. If omitted, a role is created automatically and granted read access to `s3BucketName`. Either `lambdaRoleARN` or `s3BucketName` must be provided. |
| `disabledRules` | Comma-separated list of rule IDs to disable, e.g. `"resource_lifeevent_ec2,aws_cis_3.1"`. Useful for suppressing noisy rules without redeploying code. |
| `sentryDSN` | Sentry DSN for error reporting. |

## 4. Deploy your construct

```bash
$ cdk deploy
```

# License

MIT License
