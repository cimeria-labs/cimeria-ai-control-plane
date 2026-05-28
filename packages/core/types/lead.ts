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

export interface ApolloStatusResponse {
  configured: boolean;
}

export interface ApolloSearchPreviewRequest {
  titles: string[];
  person_locations: string[];
  organization_locations: string[];
  organization_keywords: string[];
  seniorities: string[];
  limit: number;
}

export type ApolloCandidateStatus =
  | "preview"
  | "approved"
  | "enriched"
  | "imported"
  | "duplicate"
  | "rejected"
  | "missing_email"
  | "failed";

export interface ApolloCandidate {
  id: string;
  batch_id: string;
  external_id: string;
  email: string | null;
  email_status: string | null;
  name: string;
  company: string;
  title: string;
  domain: string;
  linkedin_url: string;
  status: ApolloCandidateStatus;
  score: number;
  payload: Record<string, unknown>;
}

export interface ApolloSearchPreviewResponse {
  batch_id: string;
  candidates: ApolloCandidate[];
}

export interface ApolloCandidateActionRequest {
  batch_id: string;
  candidate_ids: string[];
}

export interface ApolloEnrichResponse {
  batch_id: string;
  candidates: ApolloCandidate[];
}

export interface ApolloImportRequest {
  batch_id: string;
  candidate_ids: string[];
  no_send: true;
}

export interface ApolloImportResponse {
  batch_id: string;
  imported: number;
  skipped: number;
  missing_email: number;
  duplicates: number;
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
