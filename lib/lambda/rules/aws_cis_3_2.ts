import { cloudTrailRecord, uguisuRule, detection } from "../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.2-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.2",
      title: "AWS Management Console sign-in without MFA",
      description:
        "AWS CIS benchmark 3.2 recommend to ensure a log metric filter and alarm exist for AWS Management Console sign-in without MFA",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    if (
      record.eventName === "ConsoleLogin" &&
      record.additionalEventData !== undefined &&
      record.additionalEventData.MFAUsed !== "Yes" &&
      record.additionalEventData.SamlProviderArn === undefined
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
