import { cloudTrailRecord, uguisuRule, detection } from "../../models";

// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.7-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.7",
      title: "Disabling or scheduled deletion of customer created CMKs",
      description:
        "AWS CIS benchmark 3.7 recommend to ensure a log metric filter and alarm exist for disabling or scheduled deletion of customer created CMKs",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    if (
      record.eventSource === "kms.amazonaws.com" &&
      (record.eventName === "DisableKey" ||
        record.eventName === "ScheduleKeyDeletion")
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
