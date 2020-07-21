import { S3EventRecord, SQSEvent } from "aws-lambda";
import { S3, Lambda } from "aws-sdk";

import { gunzipSync } from "zlib";
import {
  ChatPostMessageArguments,
  MessageAttachment,
  MrkdwnElement,
} from "@slack/web-api";
import axios from "axios";

import * as models from "./models";
import * as aws_cis_3_1 from "./rules/aws_cis_3_1";
import * as aws_cis_3_2 from "./rules/aws_cis_3_2";
import * as aws_cis_3_3 from "./rules/aws_cis_3_3";
import * as aws_cis_3_4 from "./rules/aws_cis_3_4";
import * as aws_cis_3_5 from "./rules/aws_cis_3_5";
import * as aws_cis_3_6 from "./rules/aws_cis_3_6";
import * as aws_cis_3_7 from "./rules/aws_cis_3_7";
import * as aws_cis_3_8 from "./rules/aws_cis_3_8";
import * as aws_cis_3_9 from "./rules/aws_cis_3_9";
import * as aws_cis_3_10 from "./rules/aws_cis_3_10";
import * as aws_cis_3_11 from "./rules/aws_cis_3_11";
import * as aws_cis_3_12 from "./rules/aws_cis_3_12";
import * as aws_cis_3_13 from "./rules/aws_cis_3_13";
import * as aws_cis_3_14 from "./rules/aws_cis_3_14";

import * as resource_lifeevent from "./rules/resource_lifeevent";

const rules: Array<models.uguisuRule> = [
  new aws_cis_3_1.rule(),
  new aws_cis_3_2.rule(),
  new aws_cis_3_3.rule(),
  new aws_cis_3_4.rule(),
  new aws_cis_3_5.rule(),
  new aws_cis_3_6.rule(),
  new aws_cis_3_7.rule(),
  new aws_cis_3_8.rule(),
  new aws_cis_3_9.rule(),
  new aws_cis_3_10.rule(),
  new aws_cis_3_11.rule(),
  new aws_cis_3_12.rule(),
  new aws_cis_3_13.rule(),
  new aws_cis_3_14.rule(),
  new resource_lifeevent.rule(),
];

export interface arguments {
  slackWebhookURL: string;
  post(url: string, data: ChatPostMessageArguments): Promise<any>;
  getObject(params: S3.GetObjectRequest): Promise<any>;
}

export async function main(event: any, context: any) {
  console.log("context:", JSON.stringify(context));
  const s3 = new S3();

  const args = {
    slackWebhookURL: process.env.SLACK_WEBHOOK_RUL!,
    post: axios.post,
    getObject: async (params: S3.GetObjectRequest) => {
      return s3.getObject(params).promise();
    },
  };
  return handler(event, args);
}

export async function handler(event: SQSEvent, args: arguments) {
  console.log("event:", JSON.stringify(event));

  const allEvents = await fetchCloudTrailRecords(args, event);
  if (allEvents.length === 0) {
    return "no event data";
  }

  const results = allEvents
    .map((event: models.cloudTrailRecord) => {
      return rules
        .map((rule) => rule.detect(event))
        .filter((result: models.detection | null) => result !== null);
    })
    .reduce((p: Array<models.detection>, c: Array<models.detection>) => {
      return p.concat(c);
    });

  if (results.length === 0) {
    return "no detection";
  }

  console.log("detections:", JSON.stringify(results));

  const msg: ChatPostMessageArguments = {
    text: "",
    channel: "",
    attachments: results.map(buildAttachment),
  };

  console.log("slackMsg:", JSON.stringify(msg));
  const slackRes = await args.post(args.slackWebhookURL, msg);
  console.log("slackRes:", slackRes);
  return "ok";
}

function buildAttachment(log: models.detection): MessageAttachment {
  const toField = (title: string, value: string): MrkdwnElement => {
    return { type: "mrkdwn", text: "*" + title + "*\n" + value };
  };

  const ev = log.event;
  const fields = [
    toField("EventName", ev.eventName),
    toField("EventTime", ev.eventTime),
    toField("Region", ev.awsRegion),
    toField("User", ev.userIdentity ? ev.userIdentity.arn : "N/A"),
    toField("SourceIPAddress", ev.sourceIPAddress ? ev.sourceIPAddress : "N/A"),
    toField("UserAgent", ev.userAgent ? ev.userAgent : "N/A"),
  ];

  if (ev.errorCode) {
    fields.push(toField("ErrorCode", ev.errorCode));
  }
  if (ev.errorMessage) {
    fields.push(toField("ErrorMessage", ev.errorMessage));
  }

  const colorMap: { [key: string]: string } = {
    high: "danger",
    medium: "warning",
    low: "good",
  };
  const attachment: MessageAttachment = {
    color: colorMap[log.rule.severity],
    blocks: [
      {
        type: "section",
        text: { type: "mrkdwn", text: "*Detected: " + log.rule.title + "*" },
      },
      {
        type: "section",
        text: { type: "mrkdwn", text: log.rule.description },
      },
      {
        type: "section",
        fields: fields,
      },
    ],
  };

  if (ev.requestParameters) {
    const requestParameters = JSON.stringify(
      log.event.requestParameters,
      null,
      2
    );
    attachment.blocks?.push({
      type: "section",
      text: {
        type: "mrkdwn",
        text: "*RequestParameters*:\n```" + requestParameters + "```",
      },
    });
  }

  return attachment;
}

async function fetchCloudTrailRecords(args: arguments, event: SQSEvent) {
  const digestKeyPtn = new RegExp("^AWSLogs/d+/CloudTrail-Digest/");
  const s3Records = event.Records.map((record: any) => {
    const ev = JSON.parse(record.body as string);
    const msg = JSON.parse(ev.Message as string);
    return msg.Records;
  }).reduce((p: any, c: any) => {
    const i = p || [];
    const records = c.filter((r: any) => r.s3.object.key.match(digestKeyPtn));
    return i.concat(records);
  }) as Array<S3EventRecord>;

  if (!s3Records) {
    return [];
  }

  const s3proc = s3Records.map((rec) => {
    return args.getObject({
      Bucket: rec.s3.bucket.name,
      Key: rec.s3.object.key,
    });
  });

  const results = await Promise.all(s3proc);

  const allEvents = results
    .map((data) => {
      const raw = gunzipSync(data.Body as Buffer);
      const trail = JSON.parse(raw.toString());
      return trail.awsAccountId === undefined ? trail.Records : null;
    })
    .reduce((p, c) => {
      return p.concat(c);
    });

  if (!allEvents) {
    return [];
  }

  return allEvents;
}
