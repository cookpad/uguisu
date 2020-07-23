import "@aws-cdk/assert/jest";
import * as cdk from "@aws-cdk/core";
import * as Uguisu from "../lib/uguisu-stack";
import {
  arrayWith,
  objectLike,
} from "@aws-cdk/assert/lib/assertions/have-resource";

test("Build Stack", () => {
  const app = new cdk.App();
  // WHEN
  const stack = new Uguisu.UguisuStack(app, "MyTestStack", {
    lambdaRoleARN: "arn:aws:iam::1234567890:role/LambdaUguisuRole",
    slackWebhookURL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
    snsTopicARN: "arn:aws:sns:ap-northeast-1:1234567890:s3-cloudtrail-event",
  });
  // THEN
  expect(stack).toHaveResource("AWS::SQS::Queue", {
    VisibilityTimeout: 300,
  });
  expect(stack).toHaveResource("AWS::SNS::Subscription", {
    TopicArn: "arn:aws:sns:ap-northeast-1:1234567890:s3-cloudtrail-event",
    Protocol: "sqs",
  });
  expect(stack).toHaveResource("AWS::Lambda::Function", {
    Role: "arn:aws:iam::1234567890:role/LambdaUguisuRole",
    Environment: {
      Variables: {
        SLACK_WEBHOOK_RUL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
        SENTRY_DSN: "",
        DISABLE_RULES: "",
      },
    },
  });
});

test("Build Stack with S3 bucket", () => {
  const app = new cdk.App();
  // WHEN
  const stack = new Uguisu.UguisuStack(app, "MyTestStack", {
    s3BucketName: "my-test-bucket",
    slackWebhookURL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
    snsTopicARN: "arn:aws:sns:ap-northeast-1:1234567890:s3-cloudtrail-event",
  });
  // THEN
  expect(stack).toHaveResource("AWS::SQS::Queue", {
    VisibilityTimeout: 300,
  });
  expect(stack).toHaveResource("AWS::SNS::Subscription", {
    TopicArn: "arn:aws:sns:ap-northeast-1:1234567890:s3-cloudtrail-event",
    Protocol: "sqs",
  });
  expect(stack).toHaveResource("AWS::Lambda::Function", {
    Environment: {
      Variables: {
        SLACK_WEBHOOK_RUL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
        SENTRY_DSN: "",
        DISABLE_RULES: "",
      },
    },
  });

  expect(stack).toHaveResourceLike("AWS::IAM::Policy", {
    PolicyDocument: {
      Statement: arrayWith(
        objectLike({
          Action: ["s3:GetObject*", "s3:GetBucket*", "s3:List*"],
          Effect: "Allow",
          Resource: [
            "arn:aws:s3:::my-test-bucket",
            "arn:aws:s3:::my-test-bucket/*",
          ],
        })
      ),
    },
  });
});

test("Error buliding stack without S3 bucket and lambda role", () => {
  const app = new cdk.App();

  const newUguisu = () => {
    new Uguisu.UguisuStack(app, "MyTestStack", {
      slackWebhookURL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
      snsTopicARN: "arn:aws:sns:ap-northeast-1:1234567890:s3-cloudtrail-event",
    });
  };

  // WHEN
  expect(newUguisu).toThrowError();
});
