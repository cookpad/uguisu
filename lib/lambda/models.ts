export interface cloudTrailEvent {
  Records: Array<cloudTrailRecord>;
}

export interface cloudTrailRecord {
  eventTime: string;
  eventVersion: string;
  userIdentity: any;
  eventSource: string;
  eventName: string;
  awsRegion: string;
  sourceIPAddress: string;
  userAgent: string;

  errorCode?: string;
  errorMessage?: string;

  requestParameters: any;
  responseElements: any;
  additionalEventData?: any;
  requestID: string;
  eventID: string;
  eventType: string;
  apiVersion?: string;
  managementEvent?: boolean;
  readOnly?: boolean;
  resources?: any;
  recipientAccountId?: string;
  serviceEventDetails?: string;
  sharedEventID?: string;
  vpcEndpointId?: string;
  eventCategory: string;
}

export interface detection {
  rule: uguisuRule;
  event: cloudTrailRecord;
}

export interface ruleParameters {
  // Duplicated id is not allowed for each other rules.
  id: string;
  // title is short description for human.
  title: string;
  // description is more detail (What is monitored, reason of detection, etc.)
  description: string;
  // severity:
  //   high: risky event is on going
  //   medium: suspicious event
  //   low: just notification
  severity: "high" | "medium" | "low";
}

export abstract class uguisuRule {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly severity: "high" | "medium" | "low";

  constructor(params: ruleParameters) {
    this.id = params.id;
    this.title = params.title;
    this.description = params.description;
    this.severity = params.severity;
  }
  abstract detect(record: cloudTrailRecord): detection | null;
}
