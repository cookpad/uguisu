import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.6-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.6",
      title: "AWS Management Console authentication failures",
      description:
        "AWS CIS benchmark 3.6 recommend to ensure a log metric filter and alarm exist for AWS Management Console authentication failures",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    if (
      record.eventName === "ConsoleLogin" &&
      record.errorMessage === "Failed authentication"
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
