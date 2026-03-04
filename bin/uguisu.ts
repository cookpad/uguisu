#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { UguisuStack } from "../lib/uguisu-stack";
import * as path from 'path';

const app = new cdk.App();
new UguisuStack(app, process.env.UGUISU_STACK_NAME!, {
  lambdaBuildPath:  path.resolve(__dirname, '..'),
  lambdaPackagePath: './lambda/tracker',
  snsTopicARN: process.env.UGUISU_SNS_TOPIC!,
  lambdaRoleARN: process.env.UGUISU_LAMBDA_ROLE,
  s3BucketName: process.env.UGUISU_S3_BUCKET_NAME,
  slackWebhookURL: process.env.UGUISU_SLACK_WEBHOOK!,
  sentryDSN: process.env.UGUISU_SENTRY_DSN,
});
