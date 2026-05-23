import { queryOptions } from "@tanstack/react-query";
import { api } from "../api";
import type { ListLeadsParams } from "../types";

export const leadKeys = {
  all: (wsId: string) => ["leads", wsId] as const,
  list: (wsId: string, params: ListLeadsParams = {}) =>
    [...leadKeys.all(wsId), "list", params] as const,
  detail: (wsId: string, id: string) =>
    [...leadKeys.all(wsId), "detail", id] as const,
  scoreRules: (wsId: string) => [...leadKeys.all(wsId), "score-rules"] as const,
};

export function leadListOptions(wsId: string, params: ListLeadsParams = {}) {
  return queryOptions({
    queryKey: leadKeys.list(wsId, params),
    queryFn: () => api.listLeads(params),
    select: (data) => data.leads,
  });
}

export function leadDetailOptions(wsId: string, id: string) {
  return queryOptions({
    queryKey: leadKeys.detail(wsId, id),
    queryFn: () => api.getLead(id),
  });
}

export function leadScoreRuleListOptions(wsId: string) {
  return queryOptions({
    queryKey: leadKeys.scoreRules(wsId),
    queryFn: () => api.listLeadScoreRules(),
  });
}
