import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { useWorkspaceId } from "../hooks";
import type {
  CreateLeadRequest,
  Lead,
  ListLeadsResponse,
  UpdateLeadRequest,
  UpsertLeadScoreRuleRequest,
} from "../types";
import { leadKeys } from "./queries";

export function useCreateLead() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (data: CreateLeadRequest) => api.createLead(data),
    onSuccess: (lead) => {
      qc.setQueriesData<ListLeadsResponse>(
        { queryKey: leadKeys.all(wsId) },
        (old) =>
          old && "leads" in old && !old.leads.some((l) => l.id === lead.id)
            ? { ...old, leads: [lead, ...old.leads], total: old.total + 1 }
            : old,
      );
      qc.setQueryData(leadKeys.detail(wsId, lead.id), lead);
    },
    onSettled: () => {
      qc.invalidateQueries({ queryKey: leadKeys.all(wsId) });
    },
  });
}

export function useUpdateLead() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateLeadRequest) =>
      api.updateLead(id, data),
    onMutate: ({ id, ...data }) => {
      qc.cancelQueries({ queryKey: leadKeys.all(wsId) });
      const snapshots = qc.getQueriesData<ListLeadsResponse>({
        queryKey: leadKeys.all(wsId),
      });
      const detail = qc.getQueryData<Lead>(leadKeys.detail(wsId, id));

      qc.setQueriesData<ListLeadsResponse>(
        { queryKey: leadKeys.all(wsId) },
        (old) =>
          old && "leads" in old
            ? {
                ...old,
                leads: old.leads.map((lead) =>
                  lead.id === id ? { ...lead, ...data } : lead,
                ),
              }
            : old,
      );
      qc.setQueryData<Lead>(leadKeys.detail(wsId, id), (old) =>
        old ? { ...old, ...data } : old,
      );
      return { snapshots, detail, id };
    },
    onError: (_err, _vars, ctx) => {
      ctx?.snapshots.forEach(([queryKey, data]) => {
        qc.setQueryData(queryKey, data);
      });
      if (ctx?.detail) qc.setQueryData(leadKeys.detail(wsId, ctx.id), ctx.detail);
    },
    onSuccess: (lead) => {
      qc.setQueryData(leadKeys.detail(wsId, lead.id), lead);
    },
    onSettled: (_data, _err, vars) => {
      qc.invalidateQueries({ queryKey: leadKeys.all(wsId) });
      qc.invalidateQueries({ queryKey: leadKeys.detail(wsId, vars.id) });
    },
  });
}

export function useImportLeadsCsv() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (csv: string) => api.importLeadsCsv(csv),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: leadKeys.all(wsId) });
    },
  });
}

export function useUpsertLeadScoreRule() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (data: UpsertLeadScoreRuleRequest) =>
      api.upsertLeadScoreRule(data),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: leadKeys.scoreRules(wsId) });
    },
  });
}

export function useDeleteLeadScoreRule() {
  const qc = useQueryClient();
  const wsId = useWorkspaceId();
  return useMutation({
    mutationFn: (eventType: string) => api.deleteLeadScoreRule(eventType),
    onSettled: () => {
      qc.invalidateQueries({ queryKey: leadKeys.scoreRules(wsId) });
    },
  });
}
