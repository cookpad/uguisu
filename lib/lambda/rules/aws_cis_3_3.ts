import { cloudTrailRecord, uguisuRule, detection } from "../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.3-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.3",
      title: "Usage of root account",
      description:
        "AWS CIS benchmark 3.3 recommend to ensure a log metric filter and alarm exist for usage of root account",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    if (
      record.userIdentity !== undefined &&
      record.userIdentity.type === "Root" &&
      record.userIdentity.invokedBy === undefined &&
      record.eventType !== "AwsServiceEvent"
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
