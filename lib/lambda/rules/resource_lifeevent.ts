import { cloudTrailRecord, uguisuRule, detection } from "../models";

const eventMap: { [key: string]: Array<string> } = {
  "ec2.amazonaws.com": ["RunInstances", "TerminateInstances"],
  "dynamodb.amazonaws.com": ["CreateTable", "DeleteTable"],
  "cloudformation.amazonaws.com": ["CreateStack", "DeleteStack"],
  "rds.amazonaws.com": ["CreateDBInstance", "DeleteDBInstance"],
  "acm.amazonaws.com": [
    "ExportCertificate",
    "ImportCertificate",
    "RenewCertificate",
    "DeleteCertificate",
  ],
};

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent",
      title: "Resource Life Event",
      description:
        "Monitoring events of EC2, DynamoDB, CloudFormation, RDS, ACM and VPC regarding creation and destruction.",
      severity: "low",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const eventList = eventMap[record.eventSource];
    if (eventList === undefined) {
      return null;
    }

    if (eventList.indexOf(record.eventName) < 0) {
      return null;
    }

    // Ignore error event
    if (record.errorCode !== undefined) {
      return null;
    }

    // Ignore events by autoscaling or batch
    if (
      record.sourceIPAddress === "autoscaling.amazonaws.com" ||
      record.sourceIPAddress === "batch.amazonaws.com"
    ) {
      return null;
    }

    return {
      rule: this,
      event: record,
    };
  }
}
