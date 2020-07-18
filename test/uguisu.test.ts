import "@aws-cdk/assert/jest";
import * as cdk from "@aws-cdk/core";
import * as Uguisu from "../lib/uguisu-stack";

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
    Protocol: " sqs",
  });
  expect(stack).toHaveResource("AWS::Lambda::Function", {
    Role: "arn:aws:iam::1234567890:role/LambdaUguisuRole",
    Environment: {
      Variables: {
        SLACK_WEBHOOK_RUL: "https://hooks.slack.com/services/1234/5678/ABCDEFG",
      },
    },
  });
});
