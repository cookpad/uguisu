import { handler, arguments } from "../lib/lambda/tracker";
import { cloudTrailEvent, cloudTrailRecord } from "../lib/lambda/models";
import { SQSEvent } from "aws-lambda";
import { S3 } from "aws-sdk";
import { ChatPostMessageArguments } from "@slack/web-api";
import { gzipSync } from "zlib";

test("Run handler", async () => {
  const body = {
    Message: JSON.stringify({
      Records: [
        {
          s3: {
            bucket: { name: "test-bucket" },
            object: { key: "test-object-1" },
          },
        },
        {
          s3: {
            bucket: { name: "test-bucket" },
            object: { key: "test-object-2" },
          },
        },
      ],
    }),
  };

  const sqsEvent: SQSEvent = {
    Records: [
      {
        attributes: {
          SenderId: "",
          ApproximateFirstReceiveTimestamp: "",
          ApproximateReceiveCount: "",
          SentTimestamp: "",
        },
        awsRegion: "us-east-1",
        eventSource: "xxx",
        eventSourceARN: "",
        md5OfBody: "",
        messageAttributes: {},
        messageId: "",
        receiptHandle: "testhandle",
        body: JSON.stringify(body),
      },
    ],
  };

  const s3Body: { [key: string]: cloudTrailEvent } = {
    "test-object-1": {
      Records: [
        {
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "RunInstances",
          eventSource: "ec2.amazonaws.com",
          eventTime: "",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
        },
      ],
    },
    "test-object-2": {
      Records: [
        {
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "DescribeInstances",
          eventSource: "ec2.amazonaws.com",
          eventTime: "",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
        },
      ],
    },
  };

  const s3Req: Array<S3.GetObjectRequest> = [];
  const chatMsgs: Array<ChatPostMessageArguments> = [];
  const args: arguments = {
    slackWebhookURL: "https://hooks.slack.com/services/XXX/YYY/ZZZ",
    getObject: async (params: S3.GetObjectRequest) => {
      s3Req.push(params);
      const body = s3Body[params.Key];

      expect(params.Bucket).toBe("test-bucket");
      expect(body).toBeDefined();
      return {
        Body: Buffer.from(gzipSync(JSON.stringify(body))),
      };
    },
    post: async (url: string, msg: ChatPostMessageArguments) => {
      expect(url).toBe("https://hooks.slack.com/services/XXX/YYY/ZZZ");
      chatMsgs.push(msg);
    },
  };

  const result = await handler(sqsEvent, args);
  expect(result).toBe("ok");
  expect(s3Req.length).toBe(2);
  expect(chatMsgs.length).toBe(1);
});
