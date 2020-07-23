import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as sns from "@aws-cdk/aws-sns";
import * as sqs from "@aws-cdk/aws-sqs";
import * as iam from "@aws-cdk/aws-iam";
import * as s3 from "@aws-cdk/aws-s3";
import { SqsEventSource } from "@aws-cdk/aws-lambda-event-sources";
import { SqsSubscription } from "@aws-cdk/aws-sns-subscriptions";
import {
  NodejsFunction,
  NodejsFunctionProps,
} from "@aws-cdk/aws-lambda-nodejs";
import * as path from "path";

export interface Arguments {
  snsTopicARN: string;
  lambdaRoleARN?: string;
  s3BucketName?: string;
  slackWebhookURL: string;
  sentryDSN?: string;
  disableRules?: string;
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

    if (!args.lambdaRoleARN && !args.s3BucketName) {
      throw new Error(
        "Either one of lambdaRoleARN and s3BucketName is required"
      );
    }

    this.s3EventQueue = new sqs.Queue(this, "s3EventQueue", {
      visibilityTimeout: cdk.Duration.seconds(300),
    });

    const topic = sns.Topic.fromTopicArn(this, "s3Event", args.snsTopicARN);
    topic.addSubscription(new SqsSubscription(this.s3EventQueue));

    const role = args.lambdaRoleARN
      ? iam.Role.fromRoleArn(this, "Lambda", args.lambdaRoleARN, {
          mutable: false,
        })
      : undefined;

    const prop: NodejsFunctionProps = {
      entry: path.join(__dirname, "lambda/tracker.js"),
      handler: "main",
      timeout: cdk.Duration.seconds(300),
      memorySize: 1024,
      role: role,
      events: [new SqsEventSource(this.s3EventQueue, { batchSize: 1 })],
      environment: {
        SLACK_WEBHOOK_RUL: args.slackWebhookURL,
        SENTRY_DSN: args.sentryDSN || "",
        DISABLE_RULES: args.disableRules || "",
      },
    };

    this.tracker = new NodejsFunction(this, "tracker", prop);

    if (!role && args.s3BucketName) {
      const bucket = s3.Bucket.fromBucketAttributes(this, "ImportedBucket", {
        bucketArn: "arn:aws:s3:::" + args.s3BucketName,
      });
      bucket.grantRead(this.tracker);
    }
  }
}
