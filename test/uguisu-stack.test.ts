import * as cdk from "@aws-cdk/core";
import { expect as cdkExpect, haveResource, haveResourceLike, arrayWith, objectLike } from "@aws-cdk/assert";
import { UguisuStack, Arguments } from "../lib/uguisu-stack";

const baseArgs: Arguments = {
  lambdaBuildPath: ".",
  lambdaPackagePath: "./lambda/tracker",
  snsTopicARN: "arn:aws:sns:us-east-1:123456789012:my-topic",
  slackWebhookURL: "https://hooks.slack.com/services/test",
};

const testContext = {
  'aws:cdk:bundling-stacks': [],
  'aws:cdk:disable-asset-staging': true,
};

function makeStack(args: Partial<Arguments> = {}): cdk.Stack {
  const app = new cdk.App({ context: testContext });
  return new UguisuStack(app, "TestStack", { ...baseArgs, ...args });
}

test("throws when neither lambdaRoleARN nor s3BucketName is provided", () => {
  const app = new cdk.App({ context: testContext });
  expect(() => new UguisuStack(app, "TestStack", baseArgs)).toThrow(
    "Either one of lambdaRoleARN and s3BucketName is required"
  );
});

test("creates SQS queue with 300s visibility timeout", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(
    haveResource("AWS::SQS::Queue", {
      VisibilityTimeout: 300,
    })
  );
});

test("subscribes SQS queue to the SNS topic", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(haveResource("AWS::SNS::Subscription", {
    Protocol: "sqs",
    TopicArn: "arn:aws:sns:us-east-1:123456789012:my-topic",
  }));
});

test("creates Lambda with correct runtime, handler, timeout, memory, and concurrency", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::Lambda::Function", {
      Runtime: "go1.x",
      Handler: "tracker",
      Timeout: 300,
      MemorySize: 1024,
      ReservedConcurrentExecutions: 5,
    })
  );
});

test("sets SLACK_WEBHOOK_URL environment variable on Lambda", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::Lambda::Function", {
      Environment: {
        Variables: {
          SLACK_WEBHOOK_URL: "https://hooks.slack.com/services/test",
        },
      },
    })
  );
});

test("sets SENTRY_DSN to empty string when not provided", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::Lambda::Function", {
      Environment: {
        Variables: {
          SENTRY_DSN: "",
        },
      },
    })
  );
});

test("sets SENTRY_DSN when provided", () => {
  const stack = makeStack({ s3BucketName: "my-bucket", sentryDSN: "https://sentry.example.com/1" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::Lambda::Function", {
      Environment: {
        Variables: {
          SENTRY_DSN: "https://sentry.example.com/1",
        },
      },
    })
  );
});

test("uses provided IAM role ARN when lambdaRoleARN is given", () => {
  const stack = makeStack({ lambdaRoleARN: "arn:aws:iam::123456789012:role/my-role" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::Lambda::Function", {
      Role: "arn:aws:iam::123456789012:role/my-role",
    })
  );
});

test("grants S3 read to Lambda when s3BucketName is provided without a role", () => {
  const stack = makeStack({ s3BucketName: "my-bucket" });
  cdkExpect(stack).to(
    haveResourceLike("AWS::IAM::Policy", {
      PolicyDocument: {
        Statement: arrayWith(objectLike({
          Action: ["s3:GetObject*", "s3:GetBucket*", "s3:List*"],
          Effect: "Allow",
          Resource: ["arn:aws:s3:::my-bucket", "arn:aws:s3:::my-bucket/*"],
        })),
      },
    })
  );
});

test("does not create IAM policy when lambdaRoleARN is provided", () => {
  const stack = makeStack({ lambdaRoleARN: "arn:aws:iam::123456789012:role/my-role" });
  cdkExpect(stack).notTo(haveResource("AWS::IAM::Policy"));
});
