export type LeadStatus =
  | "captured"
  | "qualified"
  | "rejected"
  | "copy_ready"
  | "strategy_ready"
  | "email_sent"
  | "nurturing"
  | "hot"
  | "handoff_human"
  | "converted"
  | "cancelled";

export interface Lead {
  id: string;
  workspace_id: string;
  email: string;
  name: string;
  company: string;
  title: string;
  source: string;
  status: LeadStatus;
  score: number;
  dynamic_score: number;
  assignee_type: "member" | "agent" | null;
  assignee_id: string | null;
  pipeline_id: string | null;
  state_machine_status: string;
  last_event: string | null;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ListLeadsParams {
  limit?: number;
  offset?: number;
  status?: LeadStatus;
}

export interface ListLeadsResponse {
  leads: Lead[];
  total: number;
}

export interface CreateLeadRequest {
  email: string;
  name?: string;
  company?: string;
  title?: string;
  source?: string;
  status?: LeadStatus;
  score?: number;
  dynamic_score?: number;
  assignee_type?: "member" | "agent";
  assignee_id?: string;
  pipeline_id?: string;
  state_machine_status?: string;
  last_event?: string;
  metadata?: Record<string, unknown>;
}

export interface UpdateLeadRequest {
  email?: string;
  name?: string;
  company?: string;
  title?: string;
  source?: string;
  status?: LeadStatus;
  score?: number;
  dynamic_score?: number;
  assignee_type?: "member" | "agent" | null;
  assignee_id?: string | null;
  pipeline_id?: string | null;
  state_machine_status?: string;
  last_event?: string | null;
  metadata?: Record<string, unknown>;
}

export interface ImportLeadsResponse {
  imported: number;
  skipped: number;
  leads: Lead[];
}

export type LeadScoreEventType =
  | "opened"
  | "clicked"
  | "replied"
  | "forwarded"
  | "bounced"
  | "complained"
  | "unsubscribed";

export interface LeadScoreRule {
  id: string;
  workspace_id: string;
  event_type: LeadScoreEventType;
  weight: number;
  max_per_email: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpsertLeadScoreRuleRequest {
  event_type: LeadScoreEventType;
  weight: number;
  max_per_email?: number;
  enabled?: boolean;
}
