import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as sns from "@aws-cdk/aws-sns";
import * as sqs from "@aws-cdk/aws-sqs";
import * as iam from "@aws-cdk/aws-iam";
import { SqsEventSource } from "@aws-cdk/aws-lambda-event-sources";
import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";
import { NodejsFunction } from "@aws-cdk/aws-lambda-nodejs";
import * as path from "path";

export interface Arguments {
  snsTopicARN: string;
  lambdaRoleARN: string;
  slackWebhookURL: string;
}

export class UguisuStack extends cdk.Stack {
  s3EventQueue: sqs.Queue;
  tracker: lambda.Function;

  constructor(
    scope: cdk.Construct,
    id: string,
    args: Arguments,
    props?: cdk.StackProps
  ) {
    super(scope, id, props);

    this.s3EventQueue = new sqs.Queue(this, "s3EventQueue", {
      visibilityTimeout: cdk.Duration.seconds(300),
    });

    const topic = sns.Topic.fromTopicArn(this, "s3Event", args.snsTopicARN);
    topic.addSubscription(new SqsSubscription(this.s3EventQueue));

    const role = iam.Role.fromRoleArn(this, "Lambda", args.lambdaRoleARN, {
      mutable: false,
    });

    this.tracker = new NodejsFunction(this, "tracker", {
      entry: "lib/lambda/tracker.ts",
      handler: "main",
      timeout: cdk.Duration.seconds(300),
      role: role,
      memorySize: 1024,
      events: [new SqsEventSource(this.s3EventQueue, { batchSize: 1 })],
      environment: {
        SLACK_WEBHOOK_RUL: args.slackWebhookURL,
      },
    });
  }
}
