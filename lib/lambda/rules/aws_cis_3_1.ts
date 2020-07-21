import { cloudTrailRecord, uguisuRule, detection } from "../models";

// CIS 3.1 – Ensure a log metric filter and alarm exist for unauthorized API calls
// https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html#cis-3.1-remediation

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "aws_cis_3.1",
      title: "Unauthorized API calls monitoring",
      description:
        "AWS CIS benchmark 3.1 recommend to ensure a log metric filter and alarm exist for unauthorized API calls",
      severity: "medium",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    if (record.errorCode === undefined) {
      return null;
    }

    if (
      record.errorCode.match(/UnauthorizedOperation$/) ||
      record.errorCode === "/^AccessDenied/"
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
