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
import * as resource_lifeevent from "./rules/resource_lifeevent";

const rules: Array<models.uguisuRule> = [
  new aws_cis_3_1.rule(),
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

  const toField = (title: string, value: string): MrkdwnElement => {
    return {
      type: "mrkdwn",
      text: "*" + title + "*\n" + value,
    };
  };

  const attachments = results.map((log: models.detection) => {
    const ev = log.event;
    const fields = [
      toField("EventName", ev.eventName),
      toField("EventTime", ev.eventTime),
      toField("Region", ev.awsRegion),
      toField("User", ev.userIdentity ? ev.userIdentity.arn : "N/A"),
      toField(
        "SourceIPAddress",
        ev.sourceIPAddress ? ev.sourceIPAddress : "N/A"
      ),
      toField("UserAgent", ev.userAgent ? ev.userAgent : "N/A"),
    ];

    if (ev.errorCode) {
      fields.push(toField("ErrorCode", ev.errorCode));
    }
    if (ev.errorMessage) {
      fields.push(toField("ErrorMessage", ev.errorMessage));
    }
    if (ev.requestParameters) {
      const requestParameters = JSON.stringify(log.event.requestParameters);
      fields.push({
        type: "mrkdwn",
        text: "*RequestParameters*:\n```" + requestParameters + "```",
      });
    }

    const attachment: MessageAttachment = {
      title: "Detected: " + log.rule.title,
      text: log.rule.description,
      color: "#F2C744",
      blocks: [
        {
          type: "section",
          fields: fields,
        },
      ],
    };

    return attachment;
  });

  const msg: ChatPostMessageArguments = {
    text: "",
    channel: "",
    attachments: attachments,
  };
  const slackRes = await args.post(args.slackWebhookURL, msg);
  console.log("slackRes:", slackRes);
  return "ok";
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
