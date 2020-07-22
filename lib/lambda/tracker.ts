import { S3EventRecord, SQSEvent } from "aws-lambda";
import { S3, Lambda } from "aws-sdk";

import { gunzipSync } from "zlib";
import {
  ChatPostMessageArguments,
  MessageAttachment,
  MrkdwnElement,
  SectionBlock,
  Block,
} from "@slack/web-api";
import axios from "axios";

import * as Sentry from "@sentry/node";
Sentry.init({ dsn: process.env.SENTRY_DSN });

import * as models from "./models";
import * as aws_cis_3_1 from "./rules/aws_cis/aws_cis_3_1";
import * as aws_cis_3_2 from "./rules/aws_cis/aws_cis_3_2";
import * as aws_cis_3_3 from "./rules/aws_cis/aws_cis_3_3";
import * as aws_cis_3_4 from "./rules/aws_cis/aws_cis_3_4";
import * as aws_cis_3_5 from "./rules/aws_cis/aws_cis_3_5";
import * as aws_cis_3_6 from "./rules/aws_cis/aws_cis_3_6";
import * as aws_cis_3_7 from "./rules/aws_cis/aws_cis_3_7";
import * as aws_cis_3_8 from "./rules/aws_cis/aws_cis_3_8";
import * as aws_cis_3_9 from "./rules/aws_cis/aws_cis_3_9";
import * as aws_cis_3_10 from "./rules/aws_cis/aws_cis_3_10";
import * as aws_cis_3_11 from "./rules/aws_cis/aws_cis_3_11";
import * as aws_cis_3_12 from "./rules/aws_cis/aws_cis_3_12";
import * as aws_cis_3_13 from "./rules/aws_cis/aws_cis_3_13";
import * as aws_cis_3_14 from "./rules/aws_cis/aws_cis_3_14";

import * as resource_lifeevent_ec2 from "./rules/resource_lifeevent/ec2";
import * as resource_lifeevent_dynamodb from "./rules/resource_lifeevent/dynamodb";
import * as resource_lifeevent_rds from "./rules/resource_lifeevent/rds";
import * as resource_lifeevent_acm from "./rules/resource_lifeevent/acm";

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
  new resource_lifeevent_ec2.rule(),
  new resource_lifeevent_dynamodb.rule(),
  new resource_lifeevent_rds.rule(),
  new resource_lifeevent_acm.rule(),
];

export interface arguments {
  slackWebhookURL: string;
  disableRules?: string;

  post(url: string, data: ChatPostMessageArguments): Promise<any>;
  getObject(params: S3.GetObjectRequest): Promise<any>;
}

export async function main(event: any, context: any) {
  console.log("context:", JSON.stringify(context));
  const s3 = new S3();

  const args = {
    slackWebhookURL: process.env.SLACK_WEBHOOK_RUL!,
    disableRules: process.env.DISABLE_RULES,
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

  const disableRuleIDs: Array<string> = args.disableRules
    ? args.disableRules.split(",")
    : [];
  const enableRules = rules.filter(
    (r) => !disableRuleIDs.some((d) => d === r.id)
  );

  const results = allEvents
    .map((event: models.cloudTrailRecord) => {
      return enableRules
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

  const detections = remapDetections(results);

  const msg: ChatPostMessageArguments = {
    text: "",
    channel: "",
    attachments: Object.keys(detections).map((k) =>
      buildAttachment(detections[k])
    ),
  };

  console.log("slackMsg:", JSON.stringify(msg));
  const slackRes = await args.post(args.slackWebhookURL, msg);
  console.log("slackRes:", slackRes);
  return "ok";
}

function remapDetections(
  logs: Array<models.detection>
): { [key: string]: Array<models.detection> } {
  const dtMap: { [key: string]: Array<models.detection> } = {};

  logs.forEach((log: models.detection) => {
    if (!dtMap[log.rule.id]) {
      dtMap[log.rule.id] = [];
    }
    dtMap[log.rule.id].push(log);
  });

  return dtMap;
}

const slackColorMap: { [key: string]: string } = {
  high: "#A30200",
  medium: "#F2C744",
  low: "#2EB886",
};

function buildAttachment(logs: Array<models.detection>): MessageAttachment {
  if (logs.length === 0) {
    return { color: "#999", text: "No data" };
  }

  const toField = (title: string, value: string): MrkdwnElement => {
    return { type: "mrkdwn", text: "*" + title + "*\n" + value };
  };

  const sections = logs
    .map(
      (log: models.detection): Array<Block> => {
        const ev = log.event;
        const fields = [
          toField("EventName", ev.eventName),
          toField("EventTime", ev.eventTime),
          toField("EventID", ev.eventID),
          toField("Region", ev.awsRegion),
          toField("AccountID", ev.userIdentity.accountId),
          toField("SourceIPAddress", ev.sourceIPAddress),
          toField("User", ev.userIdentity.arn),
          toField("UserAgent", ev.userAgent),
        ];

        if (ev.errorCode) {
          fields.push(toField("ErrorCode", ev.errorCode));
        }
        if (ev.errorMessage) {
          fields.push(toField("ErrorMessage", ev.errorMessage));
        }

        const blocks: Array<SectionBlock> = [
          {
            type: "section",
            fields: fields,
          },
        ];

        if (ev.requestParameters) {
          const params = JSON.stringify(ev.requestParameters, null, 2);
          blocks.push({
            type: "section",
            text: {
              type: "mrkdwn",
              text: "*RequestParameters*:\n```" + params + "```",
            },
          });
        }

        return blocks;
      }
    )
    .reduce((p, c) => {
      if (p) {
        p.push({ type: "divider" });
      }
      return p.concat(c);
    });

  const blockHeader: Array<SectionBlock> = [
    {
      type: "section",
      text: {
        type: "mrkdwn",
        text: "*Detected: " + logs[0].rule.title + "*",
      },
    },
    {
      type: "section",
      text: { type: "mrkdwn", text: logs[0].rule.description },
    },
  ];

  const blocks: Array<Block> = [];
  const attachment: MessageAttachment = {
    color: slackColorMap[logs[0].rule.severity],
    blocks: blocks.concat(blockHeader).concat(sections),
  };

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
