#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { UguisuStack } from "../lib/uguisu-stack";

const app = new cdk.App();
new UguisuStack(app, process.env["UGUISU_STACK_NAME"]!, {
  snsTopicARN: process.env["UGUISU_SNS_TOPIC"]!,
  lambdaRoleARN: process.env["UGUISU_LAMBDA_ROLE"]!,
  slackWebhookURL: process.env["UGUISU_SLACK_WEBHOOK"]!,
});
