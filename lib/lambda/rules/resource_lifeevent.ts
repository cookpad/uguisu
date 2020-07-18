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
  "vpc.amazonaws.com": [
    "CreateRoute",
    "DeleteRoute",
    "CreateSubnet",
    "DeleteSubnet",
  ],
};

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent",
      title: "Detect resource lifecycle event",
      description:
        "Monitoring events of EC2, DynamoDB, CloudFormation, RDS, ACM and VPC regarding creation and destruction.",
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

    if (record.sourceIPAddress === "autoscaling.amazonaws.com") {
      return null;
    }

    return {
      rule: this,
      event: record,
    };
  }
}
