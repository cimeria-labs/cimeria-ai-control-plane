"use client";

import { useMemo } from "react";
import { ChevronRight, Bot } from "lucide-react";
import type { Issue } from "@multica/core/types";
import { useWorkspaceId } from "@multica/core/hooks";
import { agentListOptions } from "@multica/core/workspace/queries";
import { useQuery } from "@tanstack/react-query";

/** Generates a consistent HSL color from a string (e.g. agent ID) */
function agentColor(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 55%)`;
}

type AgentStatus = "idle" | "working" | "review";

function getAgentStatus(agentId: string, issues: Issue[]): AgentStatus {
  const agentIssues = issues.filter(
    (i) => i.assignee_type === "agent" && i.assignee_id === agentId,
  );
  if (agentIssues.some((i) => i.status === "in_review")) return "review";
  if (agentIssues.some((i) => i.status === "in_progress")) return "working";
  return "idle";
}

export function PipelineHeader({ issues }: { issues: Issue[] }) {
  const wsId = useWorkspaceId();
  const { data: agents = [] } = useQuery(agentListOptions(wsId));

  const pipeline = useMemo(() => {
    const active = agents.filter((a) => !a.archived_at);
    return active.map((agent) => ({
      id: agent.id,
      name: agent.name,
      status: getAgentStatus(agent.id, issues),
      issueCount: issues.filter(
        (i) => i.assignee_type === "agent" && i.assignee_id === agent.id,
      ).length,
    }));
  }, [agents, issues]);

  if (pipeline.length === 0) return null;

  return (
    <div className="flex items-center gap-1.5 overflow-x-auto px-4 py-2 text-xs border-b bg-muted/20">
      <Bot className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
      <span className="text-muted-foreground font-medium shrink-0">Pipeline</span>
      <div className="flex items-center gap-1 ml-1">
        {pipeline.map((agent, i) => (
          <PipelineNode key={agent.id} agent={agent} showArrow={i > 0} />
        ))}
      </div>
    </div>
  );
}

function PipelineNode({
  agent,
  showArrow,
}: {
  agent: { id: string; name: string; status: AgentStatus; issueCount: number };
  showArrow: boolean;
}) {
  const color = agentColor(agent.id);
  const dotClass =
    agent.status === "working"
      ? "animate-pulse"
      : agent.status === "review"
        ? "animate-approval-dot"
        : "";

  const statusColor =
    agent.status === "working"
      ? "var(--warning)"
      : agent.status === "review"
        ? "var(--success)"
        : "var(--muted-foreground)";

  return (
    <>
      {showArrow && (
        <ChevronRight className="h-3 w-3 text-muted-foreground/50 shrink-0" />
      )}
      <div
        className="flex items-center gap-1.5 rounded-md px-2 py-1 bg-background border transition-colors"
        style={{ borderColor: agent.status !== "idle" ? color : undefined }}
      >
        <span
          className={`inline-block h-1.5 w-1.5 rounded-full shrink-0 ${dotClass}`}
          style={{ backgroundColor: statusColor }}
        />
        <span className="truncate max-w-[100px] font-medium">{agent.name}</span>
        {agent.issueCount > 0 && (
          <span className="text-muted-foreground tabular-nums">
            {agent.issueCount}
          </span>
        )}
      </div>
    </>
  );
}