import { cloudTrailRecord, uguisuRule, detection } from "../../models";

export class rule extends uguisuRule {
  constructor() {
    super({
      id: "resource_lifeevent_rds",
      title: "RDS Life Event",
      description: "Monitoring events of RDS creation and destruction.",
      severity: "low",
    });
  }

  detect(record: cloudTrailRecord): detection | null {
    const targetEvents = new Set(["CreateDBInstance", "DeleteDBInstance"]);

    if (
      record.eventSource === "rds.amazonaws.com" &&
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
