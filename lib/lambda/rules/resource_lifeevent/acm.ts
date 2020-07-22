import { cloudTrailRecord, uguisuRule, detection } from "../../models";

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent_acm",
      title: "ACM Life Event",
      description: "Monitoring events of ACM creation and destruction.",
      severity: "low",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents = new Set([
      "ExportCertificate",
      "ImportCertificate",
      "RenewCertificate",
      "DeleteCertificate",
    ]);

    if (
      record.eventSource === "acm.amazonaws.com" &&
      targetEvents.has(record.eventName)
    ) {
      return {
        rule: this,
        event: record,
      };
    }

    return null;
  }
}
