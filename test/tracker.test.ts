import { handler, arguments } from "../lib/lambda/tracker";
import { cloudTrailEvent, cloudTrailRecord } from "../lib/lambda/models";
import { SQSEvent } from "aws-lambda";
import { S3 } from "aws-sdk";
import { ChatPostMessageArguments } from "@slack/web-api";
import { gzipSync } from "zlib";

test("Handler basic", async () => {
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
          eventTime: "2020-04-20T12:34:56",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
          requestParameters: {
            instanceType: "c4.8xlarge",
          },
          responseElements: {},
          eventCategory: "",
          sourceIPAddress: "",
          userAgent: "",
          userIdentity: {},
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
          responseElements: {},
          eventCategory: "",
          sourceIPAddress: "",
          userAgent: "",
          userIdentity: {},
          requestParameters: {},
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
  const msg = JSON.parse(JSON.stringify(chatMsgs[0]));
  expect(msg.attachments).toBeDefined();
  expect(msg.attachments.length).toBe(1);
  expect(msg.attachments[0].blocks).toBeDefined();
  expect(msg.attachments[0].blocks.length).toBe(4);
  expect(msg.attachments[0].blocks[0].text.text).toContain(
    "Detected: EC2 Life Event"
  );
  expect(msg.attachments[0].blocks[2].fields).toBeDefined();
  const fields = msg.attachments[0].blocks[2].fields;
  expect(fields[0].text).toContain("*EventName*");
  expect(fields[0].text).toContain("RunInstances");
  expect(fields[1].text).toContain("*EventTime*");
  expect(fields[1].text).toContain("2020-04-20T12:34:56");
});

describe("handler variation", async () => {
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
        body: JSON.stringify({
          Message: JSON.stringify({
            Records: [
              {
                s3: {
                  bucket: { name: "test-bucket" },
                  object: { key: "test-object" },
                },
              },
            ],
          }),
        }),
      },
    ],
  };

  const dummyURL = "https://hooks.slack.com/services/XXX/YYY/ZZZ";

  const newGetObject = (body: any) => {
    return async (params: S3.GetObjectRequest) => {
      return { Body: Buffer.from(gzipSync(JSON.stringify(body))) };
    };
  };
  const newPost = (chatMsgs: Array<ChatPostMessageArguments>) => {
    return async (url: string, msg: ChatPostMessageArguments) => {
      chatMsgs.push(msg);
    };
  };

  it("simple run", async () => {
    const body = {
      Records: [
        {
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "RunInstances",
          eventSource: "ec2.amazonaws.com",
          eventTime: "2020-04-20T12:34:56",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
          requestParameters: {
            instanceType: "c4.8xlarge",
          },
        },
      ],
    };

    const chatMsgs: Array<ChatPostMessageArguments> = [];
    const args: arguments = {
      slackWebhookURL: dummyURL,
      getObject: newGetObject(body),
      post: newPost(chatMsgs),
    };

    const result = await handler(sqsEvent, args);
    expect(result).toBe("ok");
    expect(chatMsgs.length).toBe(1);
    const msg = JSON.parse(JSON.stringify(chatMsgs[0]));
    expect(msg.attachments[0].blocks[0].text.text).toContain(
      "Detected: EC2 Life Event"
    );
  });

  describe("disable rule", () => {
    const body: cloudTrailEvent = {
      Records: [
        {
          // aws_cis_3.4
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "CreatePolicy",
          eventSource: "",
          eventTime: "2020-04-20T12:34:56",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
          requestParameters: {},
          responseElements: {},
          eventCategory: "",
          sourceIPAddress: "",
          userAgent: "",
          userIdentity: {},
        },
        {
          // aws_cis_3.5
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "CreateTrail",
          eventSource: "",
          eventTime: "2020-04-20T12:34:56",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
          requestParameters: {},
          responseElements: {},
          eventCategory: "",
          sourceIPAddress: "",
          userAgent: "",
          userIdentity: {},
        },
        {
          // aws_cis_3.6
          awsRegion: "ap-northeast-1",
          eventID: "x123",
          eventName: "ConsoleLogin",
          eventSource: "",
          eventTime: "2020-04-20T12:34:56",
          eventType: "",
          eventVersion: "",
          recipientAccountId: "",
          requestID: "",
          requestParameters: {},
          errorMessage: "Failed authentication",
          responseElements: {},
          eventCategory: "",
          sourceIPAddress: "",
          userAgent: "",
          userIdentity: {},
        },
      ],
    };

    it("disable 2 rules", async () => {
      const chatMsgs: Array<ChatPostMessageArguments> = [];
      const args: arguments = {
        slackWebhookURL: dummyURL,
        disableRules: "aws_cis_3.4,aws_cis_3.6",
        getObject: newGetObject(body),
        post: newPost(chatMsgs),
      };

      const result = await handler(sqsEvent, args);
      expect(result).toBe("ok");
      expect(chatMsgs.length).toBe(1);
      expect(chatMsgs[0].attachments!.length).toBe(1);
      const msg = JSON.parse(JSON.stringify(chatMsgs[0]));
      expect(msg.attachments[0].blocks[0].text.text).toContain(
        "Detected: CloudTrail configuration changes"
      );
    });

    it("disable 1 rule", async () => {
      const chatMsgs: Array<ChatPostMessageArguments> = [];
      const args: arguments = {
        slackWebhookURL: dummyURL,
        disableRules: "aws_cis_3.5",
        getObject: newGetObject(body),
        post: newPost(chatMsgs),
      };

      const result = await handler(sqsEvent, args);
      expect(result).toBe("ok");
      expect(chatMsgs.length).toBe(1);
      expect(chatMsgs[0].attachments!.length).toBe(2);
      const msg = JSON.parse(JSON.stringify(chatMsgs[0]));
      expect(msg.attachments[0].blocks[0].text.text).toContain(
        "Detected: IAM policy changes"
      );
      expect(msg.attachments[1].blocks[0].text.text).toContain(
        "Detected: AWS Management Console authentication failures"
      );
    });
  });
});
