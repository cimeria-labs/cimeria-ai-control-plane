"use client";

import { useCallback, memo } from "react";
import { AppLink } from "../../navigation";
import { useSortable, defaultAnimateLayoutChanges } from "@dnd-kit/sortable";
import type { AnimateLayoutChanges } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { toast } from "sonner";
import type { Issue, UpdateIssueRequest } from "@multica/core/types";
import { CalendarDays, CheckCircle2, Bot } from "lucide-react";
import { ActorAvatar } from "../../common/actor-avatar";
import { useActorName } from "@multica/core/workspace/hooks";

/** Generates a consistent HSL color from a string (e.g. agent ID) */
function agentRingColor(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 70%, 55%)`;
}
import { useUpdateIssue } from "@multica/core/issues/mutations";
import { useWorkspacePaths } from "@multica/core/paths";
import { PriorityIcon } from "./priority-icon";
import { PriorityPicker, AssigneePicker, DueDatePicker } from "./pickers";
import { PRIORITY_CONFIG } from "@multica/core/issues/config";
import { useViewStore } from "@multica/core/issues/stores/view-store-context";
import { ProgressRing } from "./progress-ring";
import type { ChildProgress } from "./list-row";

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

/** Stops event from bubbling to Link/drag handlers */
function PickerWrapper({ children }: { children: React.ReactNode }) {
  const stop = (e: React.SyntheticEvent) => {
    e.stopPropagation();
    e.preventDefault();
  };
  return (
    <div onClick={stop} onMouseDown={stop} onPointerDown={stop}>
      {children}
    </div>
  );
}

export const BoardCardContent = memo(function BoardCardContent({
  issue,
  editable = false,
  childProgress,
}: {
  issue: Issue;
  editable?: boolean;
  childProgress?: ChildProgress;
}) {
  const storeProperties = useViewStore((s) => s.cardProperties);
  const priorityCfg = PRIORITY_CONFIG[issue.priority];

  const updateIssueMutation = useUpdateIssue();
  const handleUpdate = useCallback(
    (updates: Partial<UpdateIssueRequest>) => {
      updateIssueMutation.mutate(
        { id: issue.id, ...updates },
        { onError: () => toast.error("Failed to update issue") },
      );
    },
    [issue.id, updateIssueMutation],
  );

  const showPriority = storeProperties.priority;
  const showDescription = storeProperties.description && issue.description;
  const showAssignee = storeProperties.assignee && issue.assignee_type && issue.assignee_id;
  const showDueDate = storeProperties.dueDate && issue.due_date;

  const isAgentWorking = issue.assignee_type === "agent" && issue.status === "in_progress";
  const isAwaitingReview = issue.status === "in_review";

  const { getActorName } = useActorName();
  const isAgentAssignee = !!showAssignee && issue.assignee_type === "agent";
  const agentName = isAgentAssignee ? getActorName("agent", issue.assignee_id!) : null;
  const agentColor = isAgentAssignee ? agentRingColor(issue.assignee_id!) : null;

  return (
    <div className={`rounded-lg border bg-card p-3.5 shadow-[0_1px_2px_0_rgba(0,0,0,0.03)] transition-shadow group-hover:shadow-sm ${
      isAwaitingReview
        ? "border-success/40 animate-approval-glow"
        : isAgentWorking
          ? "border-warning/30 animate-agent-pulse"
          : ""
    }`}>
      {/* Row 1: Identifier + agent badge */}
      <div className="flex items-center gap-1.5">
        <p className="text-xs text-muted-foreground">{issue.identifier}</p>
        {isAgentWorking && (
          <span className="inline-flex items-center gap-1 rounded-full bg-warning/10 px-1.5 py-0.5 text-[10px] font-medium text-warning">
            <Bot className="h-2.5 w-2.5 animate-pulse" />
            Working
          </span>
        )}
        {isAwaitingReview && (
          <span className="inline-flex items-center gap-1 rounded-full bg-success/10 px-1.5 py-0.5 text-[10px] font-medium text-success">
            <CheckCircle2 className="h-2.5 w-2.5" />
            Review
          </span>
        )}
      </div>

      {/* Row 2: Title */}
      <p className="mt-1 text-sm font-medium leading-snug line-clamp-2">
        {issue.title}
      </p>

      {/* Approval button for in_review issues */}
      {isAwaitingReview && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            e.preventDefault();
            handleUpdate({ status: "done" });
            toast.success("Issue approved");
          }}
          className="mt-2 flex w-full items-center justify-center gap-1.5 rounded-md bg-success/10 px-2 py-1.5 text-xs font-semibold text-success hover:bg-success/20 transition-colors"
        >
          <CheckCircle2 className="h-3.5 w-3.5" />
          Approve
        </button>
      )}

      {/* Sub-issue progress */}
      {childProgress && (
        <div className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-muted/60 px-1.5 py-0.5">
          <ProgressRing done={childProgress.done} total={childProgress.total} size={14} />
          <span className="text-[11px] text-muted-foreground tabular-nums font-medium">
            {childProgress.done}/{childProgress.total}
          </span>
        </div>
      )}

      {/* Description */}
      {showDescription && (
        <p className="mt-1 text-xs text-muted-foreground line-clamp-1">
          {issue.description}
        </p>
      )}

      {/* Row 3: Assignee, priority badge, due date */}
      {(showAssignee || showPriority || showDueDate) && (
        <div className="mt-3 flex items-center gap-2">
          {showAssignee &&
            (editable ? (
              <PickerWrapper>
                <AssigneePicker
                  assigneeType={issue.assignee_type}
                  assigneeId={issue.assignee_id}
                  onUpdate={handleUpdate}
                  trigger={
                    <div className="flex items-center gap-1">
                      <span
                        className="inline-flex rounded-full border-2 p-0.5"
                        style={{ borderColor: issue.assignee_type === "agent" ? agentRingColor(issue.assignee_id!) : "transparent" }}
                      >
                        <ActorAvatar
                          actorType={issue.assignee_type!}
                          actorId={issue.assignee_id!}
                          size={18}
                        />
                      </span>
                      {agentName && (
                        <span className="truncate max-w-[64px] text-[10px] font-medium" style={{ color: agentColor! }}>
                          {agentName}
                        </span>
                      )}
                    </div>
                  }
                />
              </PickerWrapper>
            ) : (
              <div className="flex items-center gap-1">
                <span
                  className="inline-flex rounded-full border-2 p-0.5"
                  style={{ borderColor: issue.assignee_type === "agent" ? agentRingColor(issue.assignee_id!) : "transparent" }}
                >
                  <ActorAvatar
                    actorType={issue.assignee_type!}
                    actorId={issue.assignee_id!}
                    size={18}
                  />
                </span>
                {agentName && (
                  <span className="truncate max-w-[64px] text-[10px] font-medium" style={{ color: agentColor! }}>
                    {agentName}
                  </span>
                )}
              </div>
            ))}
          {showPriority &&
            (editable ? (
              <PickerWrapper>
                <PriorityPicker
                  priority={issue.priority}
                  onUpdate={handleUpdate}
                  trigger={
                    <span className={`inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs font-medium ${priorityCfg.badgeBg} ${priorityCfg.badgeText}`}>
                      <PriorityIcon priority={issue.priority} className="h-3 w-3" inheritColor />
                      {priorityCfg.label}
                    </span>
                  }
                />
              </PickerWrapper>
            ) : (
              <span className={`inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs font-medium ${priorityCfg.badgeBg} ${priorityCfg.badgeText}`}>
                <PriorityIcon priority={issue.priority} className="h-3 w-3" inheritColor />
                {priorityCfg.label}
              </span>
            ))}
          {showDueDate && (
            <div className="ml-auto">
              {editable ? (
                <PickerWrapper>
                  <DueDatePicker
                    dueDate={issue.due_date}
                    onUpdate={handleUpdate}
                    trigger={
                      <span
                        className={`flex items-center gap-1 text-xs ${
                          new Date(issue.due_date!) < new Date()
                            ? "text-destructive"
                            : "text-muted-foreground"
                        }`}
                      >
                        <CalendarDays className="size-3" />
                        {formatDate(issue.due_date!)}
                      </span>
                    }
                  />
                </PickerWrapper>
              ) : (
                <span
                  className={`flex items-center gap-1 text-xs ${
                    new Date(issue.due_date!) < new Date()
                      ? "text-destructive"
                      : "text-muted-foreground"
                  }`}
                >
                  <CalendarDays className="size-3" />
                  {formatDate(issue.due_date!)}
                </span>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
});

const animateLayoutChanges: AnimateLayoutChanges = (args) => {
  const { isSorting, wasDragging } = args;
  if (isSorting || wasDragging) return false;
  return defaultAnimateLayoutChanges(args);
};

export const DraggableBoardCard = memo(function DraggableBoardCard({ issue, childProgress }: { issue: Issue; childProgress?: ChildProgress }) {
  const p = useWorkspacePaths();
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: issue.id,
    data: { status: issue.status },
    animateLayoutChanges,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={isDragging ? "opacity-30" : ""}
    >
      <AppLink
        href={p.issueDetail(issue.id)}
        className={`group block transition-colors ${isDragging ? "pointer-events-none" : ""}`}
      >
        <BoardCardContent issue={issue} editable childProgress={childProgress} />
      </AppLink>
    </div>
  );
});
