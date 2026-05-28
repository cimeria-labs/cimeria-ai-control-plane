"use client";

import { useMemo, useState, type FormEvent } from "react";
import { toast } from "sonner";
import { Check, ChevronRight, Flame, Plus, Search, ShieldCheck, Sparkles, Upload, UsersRound } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { apolloStatusOptions, leadListOptions } from "@multica/core/leads/queries";
import {
  useCreateLead,
  useEnrichApolloCandidates,
  useImportApprovedApolloCandidates,
  useImportLeadsCsv,
  useSearchApolloPreview,
  useUpdateLead,
} from "@multica/core/leads/mutations";
import { useCurrentWorkspace } from "@multica/core/paths";
import { useWorkspaceId } from "@multica/core/hooks";
import type { ApolloCandidate, Lead, LeadStatus } from "@multica/core/types";
import { Button } from "@multica/ui/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@multica/ui/components/ui/dialog";
import { Input } from "@multica/ui/components/ui/input";
import { NativeSelect, NativeSelectOption } from "@multica/ui/components/ui/native-select";
import { Skeleton } from "@multica/ui/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@multica/ui/components/ui/table";
import { Textarea } from "@multica/ui/components/ui/textarea";
import { cn } from "@multica/ui/lib/utils";
import { PageHeader } from "../../layout/page-header";
import { WorkspaceAvatar } from "../../workspace/workspace-avatar";

const STATUS_ORDER: LeadStatus[] = [
  "captured",
  "qualified",
  "rejected",
  "copy_ready",
  "strategy_ready",
  "email_sent",
  "nurturing",
  "hot",
  "handoff_human",
  "converted",
  "cancelled",
];

const STATUS_CONFIG: Record<LeadStatus, { label: string; className: string }> = {
  captured: { label: "Captured", className: "bg-slate-100 text-slate-700 dark:bg-slate-900/50 dark:text-slate-300" },
  qualified: { label: "Qualified", className: "bg-sky-100 text-sky-700 dark:bg-sky-950/50 dark:text-sky-300" },
  rejected: { label: "Rejected", className: "bg-rose-100 text-rose-700 dark:bg-rose-950/50 dark:text-rose-300" },
  copy_ready: { label: "Copy ready", className: "bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300" },
  strategy_ready: { label: "Strategy ready", className: "bg-cyan-100 text-cyan-700 dark:bg-cyan-950/50 dark:text-cyan-300" },
  email_sent: { label: "Email sent", className: "bg-indigo-100 text-indigo-700 dark:bg-indigo-950/50 dark:text-indigo-300" },
  nurturing: { label: "Nurturing", className: "bg-amber-100 text-amber-800 dark:bg-amber-950/50 dark:text-amber-300" },
  hot: { label: "Hot", className: "bg-red-100 text-red-700 dark:bg-red-950/50 dark:text-red-300" },
  handoff_human: { label: "Handoff", className: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/50 dark:text-emerald-300" },
  converted: { label: "Converted", className: "bg-green-100 text-green-700 dark:bg-green-950/50 dark:text-green-300" },
  cancelled: { label: "Cancelled", className: "bg-muted text-muted-foreground" },
};

const DEFAULT_APOLLO_TITLES = "Founder, CEO, Head of AI";
const DEFAULT_APOLLO_LOCATIONS = "Brazil";
const DEFAULT_APOLLO_KEYWORDS = "artificial intelligence, inteligencia artificial, AI";
const DEFAULT_APOLLO_SENIORITIES = "founder, owner, c_suite, vp, director, head";

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
  }).format(new Date(value));
}

function displayName(lead: Lead) {
  return lead.name.trim() || lead.email;
}

function totalScore(lead: Lead) {
  return lead.score + lead.dynamic_score;
}

function splitList(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function mergeCandidates(current: ApolloCandidate[], next: ApolloCandidate[]) {
  const byID = new Map(current.map((candidate) => [candidate.id, candidate]));
  for (const candidate of next) byID.set(candidate.id, candidate);
  return Array.from(byID.values());
}

function StatusBadge({ status }: { status: LeadStatus }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <span className={cn("inline-flex h-6 items-center rounded-md px-2 text-xs font-medium", cfg.className)}>
      {status === "hot" && <Flame className="mr-1 size-3" />}
      {cfg.label}
    </span>
  );
}

function LeadScoreInput({ lead }: { lead: Lead }) {
  const updateLead = useUpdateLead();
  return (
    <input
      type="number"
      min={0}
      max={100}
      defaultValue={lead.score}
      onBlur={(event) => {
        const next = Number(event.currentTarget.value);
        if (Number.isNaN(next) || next === lead.score) return;
        updateLead.mutate(
          { id: lead.id, score: next },
          { onError: () => toast.error("Failed to update score") },
        );
      }}
      className="h-7 w-14 rounded-md border bg-transparent px-2 text-right text-xs tabular-nums outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/40"
    />
  );
}

function CreateLeadDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const createLead = useCreateLead();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [company, setCompany] = useState("");
  const [title, setTitle] = useState("");
  const [source, setSource] = useState("manual");

  const reset = () => {
    setEmail("");
    setName("");
    setCompany("");
    setTitle("");
    setSource("manual");
  };

  const submit = (event: FormEvent) => {
    event.preventDefault();
    createLead.mutate(
      {
        email,
        name: name || undefined,
        company: company || undefined,
        title: title || undefined,
        source: source || undefined,
      },
      {
        onSuccess: () => {
          toast.success("Lead created");
          reset();
          onOpenChange(false);
        },
        onError: (error) => toast.error(error instanceof Error ? error.message : "Failed to create lead"),
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New lead</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="grid gap-3">
          <Input value={email} onChange={(e) => setEmail(e.target.value)} type="email" placeholder="email@company.com" required />
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" />
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Input value={company} onChange={(e) => setCompany(e.target.value)} placeholder="Company" />
            <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Title" />
          </div>
          <Input value={source} onChange={(e) => setSource(e.target.value)} placeholder="Source" />
          <DialogFooter className="mt-1">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={createLead.isPending}>
              <Plus />
              Create
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function ImportCsvDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const importCsv = useImportLeadsCsv();
  const [csv, setCsv] = useState("email,name,company,title,source\n");

  const submit = (event: FormEvent) => {
    event.preventDefault();
    importCsv.mutate(csv, {
      onSuccess: (result) => {
        toast.success(`Imported ${result.imported} leads`);
        setCsv("email,name,company,title,source\n");
        onOpenChange(false);
      },
      onError: (error) => toast.error(error instanceof Error ? error.message : "Failed to import CSV"),
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Import CSV</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="grid gap-3">
          <Textarea
            value={csv}
            onChange={(e) => setCsv(e.target.value)}
            className="min-h-52 font-mono text-xs"
            spellCheck={false}
          />
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
            <Button type="submit" disabled={importCsv.isPending}>
              <Upload />
              Import
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function ApolloImportDialog({
  open,
  onOpenChange,
  configured,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  configured: boolean;
}) {
  const searchApollo = useSearchApolloPreview();
  const enrichApollo = useEnrichApolloCandidates();
  const importApollo = useImportApprovedApolloCandidates();
  const [titles, setTitles] = useState(DEFAULT_APOLLO_TITLES);
  const [locations, setLocations] = useState(DEFAULT_APOLLO_LOCATIONS);
  const [keywords, setKeywords] = useState(DEFAULT_APOLLO_KEYWORDS);
  const [seniorities, setSeniorities] = useState(DEFAULT_APOLLO_SENIORITIES);
  const [limit, setLimit] = useState(10);
  const [batchID, setBatchID] = useState("");
  const [candidates, setCandidates] = useState<ApolloCandidate[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const selectedIDs = useMemo(
    () => candidates.filter((candidate) => selected.has(candidate.id)).map((candidate) => candidate.id),
    [candidates, selected],
  );
  const enrichedCount = candidates.filter((candidate) => Boolean(candidate.email)).length;
  const busy = searchApollo.isPending || enrichApollo.isPending || importApollo.isPending;

  const reset = () => {
    setBatchID("");
    setCandidates([]);
    setSelected(new Set());
  };

  const toggleCandidate = (id: string) => {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const search = (event: FormEvent) => {
    event.preventDefault();
    if (!configured) return;
    searchApollo.mutate(
      {
        titles: splitList(titles),
        person_locations: [],
        organization_locations: splitList(locations),
        organization_keywords: splitList(keywords),
        seniorities: splitList(seniorities),
        limit,
      },
      {
        onSuccess: (result) => {
          setBatchID(result.batch_id);
          setCandidates(result.candidates);
          setSelected(new Set(result.candidates.map((candidate) => candidate.id)));
          toast.success(`Apollo returned ${result.candidates.length} candidates`);
        },
        onError: (error) => toast.error(error instanceof Error ? error.message : "Apollo search failed"),
      },
    );
  };

  const enrich = () => {
    if (!batchID || selectedIDs.length === 0) return;
    enrichApollo.mutate(
      { batch_id: batchID, candidate_ids: selectedIDs },
      {
        onSuccess: (result) => {
          setCandidates((current) => mergeCandidates(current, result.candidates));
          toast.success(`Enriched ${result.candidates.length} candidates`);
        },
        onError: (error) => toast.error(error instanceof Error ? error.message : "Apollo enrichment failed"),
      },
    );
  };

  const importSelected = () => {
    if (!batchID || selectedIDs.length === 0) return;
    importApollo.mutate(
      { batch_id: batchID, candidate_ids: selectedIDs, no_send: true },
      {
        onSuccess: (result) => {
          toast.success(`Imported ${result.imported} Apollo leads`);
          reset();
          onOpenChange(false);
        },
        onError: (error) => toast.error(error instanceof Error ? error.message : "Apollo import failed"),
      },
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl">
        <DialogHeader>
          <div className="flex items-center justify-between gap-3">
            <DialogTitle>Apollo</DialogTitle>
            <span className="inline-flex h-6 items-center gap-1.5 rounded-md bg-emerald-50 px-2 text-xs font-medium text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300">
              <ShieldCheck className="size-3" />
              No send
            </span>
          </div>
        </DialogHeader>

        <form onSubmit={search} className="grid gap-3">
          <div className="grid gap-3 md:grid-cols-[1.2fr_1fr]">
            <Input value={titles} onChange={(e) => setTitles(e.target.value)} placeholder="Titles" disabled={!configured || busy} />
            <Input value={locations} onChange={(e) => setLocations(e.target.value)} placeholder="Company locations" disabled={!configured || busy} />
            <Input value={keywords} onChange={(e) => setKeywords(e.target.value)} placeholder="Company keywords" disabled={!configured || busy} />
            <div className="grid grid-cols-[1fr_88px] gap-3">
              <Input value={seniorities} onChange={(e) => setSeniorities(e.target.value)} placeholder="Seniorities" disabled={!configured || busy} />
              <Input
                value={limit}
                onChange={(e) => setLimit(Math.max(1, Math.min(10, Number(e.target.value) || 1)))}
                type="number"
                min={1}
                max={10}
                aria-label="Apollo limit"
                disabled={!configured || busy}
              />
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button type="submit" size="sm" disabled={!configured || busy}>
              <Search />
              Search
            </Button>
            <Button type="button" size="sm" variant="outline" onClick={enrich} disabled={!configured || busy || selectedIDs.length === 0}>
              <Sparkles />
              Enrich
            </Button>
            <Button type="button" size="sm" onClick={importSelected} disabled={!configured || busy || selectedIDs.length === 0}>
              <ShieldCheck />
              Import
            </Button>
            <span className="ml-auto text-xs text-muted-foreground">
              {selectedIDs.length} selected / {enrichedCount} enriched
            </span>
          </div>
        </form>

        <div className="max-h-[48vh] overflow-auto rounded-md border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-10" />
                <TableHead>Contact</TableHead>
                <TableHead>Company</TableHead>
                <TableHead>Email</TableHead>
                <TableHead className="text-right">Score</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {candidates.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="h-24 text-center text-sm text-muted-foreground">
                    {configured ? "No candidates" : "Apollo not configured"}
                  </TableCell>
                </TableRow>
              ) : (
                candidates.map((candidate) => (
                  <TableRow key={candidate.id}>
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selected.has(candidate.id)}
                        onChange={() => toggleCandidate(candidate.id)}
                        aria-label={`Select ${candidate.name || candidate.company || "candidate"}`}
                        className="size-4 rounded border-input accent-primary"
                      />
                    </TableCell>
                    <TableCell>
                      <div className="min-w-0">
                        <div className="truncate font-medium">{candidate.name || "--"}</div>
                        <div className="truncate text-xs text-muted-foreground">{candidate.title || "--"}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="min-w-0">
                        <div className="truncate">{candidate.company || "--"}</div>
                        <div className="truncate text-xs text-muted-foreground">{candidate.domain || "--"}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="min-w-0">
                        <div className="truncate">{candidate.email || "--"}</div>
                        <div className="truncate text-xs text-muted-foreground">{candidate.email_status || candidate.status}</div>
                      </div>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">{candidate.score}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function LeadsTable({ leads }: { leads: Lead[] }) {
  const updateLead = useUpdateLead();

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="w-[30%] pl-5">Lead</TableHead>
          <TableHead>Company</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Base</TableHead>
          <TableHead className="text-right">Dynamic</TableHead>
          <TableHead className="text-right">Total</TableHead>
          <TableHead>Source</TableHead>
          <TableHead className="pr-5 text-right">Updated</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {leads.map((lead) => (
          <TableRow key={lead.id}>
            <TableCell className="pl-5">
              <div className="min-w-0">
                <div className="truncate font-medium">{displayName(lead)}</div>
                <div className="truncate text-xs text-muted-foreground">{lead.email}</div>
              </div>
            </TableCell>
            <TableCell>
              <div className="min-w-0">
                <div className="truncate">{lead.company || "--"}</div>
                <div className="truncate text-xs text-muted-foreground">{lead.title || "--"}</div>
              </div>
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-2">
                <StatusBadge status={lead.status} />
                <NativeSelect
                  size="sm"
                  value={lead.status}
                  aria-label="Lead status"
                  onChange={(event) => {
                    updateLead.mutate(
                      { id: lead.id, status: event.currentTarget.value as LeadStatus },
                      { onError: () => toast.error("Failed to update status") },
                    );
                  }}
                  className="w-36"
                >
                  {STATUS_ORDER.map((status) => (
                    <NativeSelectOption key={status} value={status}>
                      {STATUS_CONFIG[status].label}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </div>
            </TableCell>
            <TableCell className="text-right"><LeadScoreInput lead={lead} /></TableCell>
            <TableCell className="text-right tabular-nums">{lead.dynamic_score}</TableCell>
            <TableCell className={cn("text-right font-medium tabular-nums", totalScore(lead) >= 7 && "text-red-600")}>
              {totalScore(lead)}
            </TableCell>
            <TableCell className="text-muted-foreground">{lead.source || "manual"}</TableCell>
            <TableCell className="pr-5 text-right text-muted-foreground">{formatDate(lead.updated_at)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export function LeadsPage() {
  const wsId = useWorkspaceId();
  const workspace = useCurrentWorkspace();
  const [statusFilter, setStatusFilter] = useState<LeadStatus | "all">("all");
  const [createOpen, setCreateOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [apolloOpen, setApolloOpen] = useState(false);
  const params = useMemo(
    () => (statusFilter === "all" ? {} : { status: statusFilter }),
    [statusFilter],
  );
  const { data: leads = [], isLoading } = useQuery(leadListOptions(wsId, params));
  const { data: apolloStatus } = useQuery(apolloStatusOptions(wsId));
  const apolloConfigured = Boolean(apolloStatus?.configured);
  const hotCount = leads.filter((lead) => totalScore(lead) >= 7 || lead.status === "hot").length;

  if (isLoading) {
    return (
      <div className="flex flex-1 min-h-0 flex-col">
        <PageHeader className="justify-between">
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-8 w-28" />
        </PageHeader>
        <div className="flex h-11 items-center gap-2 border-b px-5">
          {Array.from({ length: 5 }).map((_, index) => (
            <Skeleton key={index} className="h-7 w-24 rounded-md" />
          ))}
        </div>
        <div className="space-y-1 p-5">
          {Array.from({ length: 8 }).map((_, index) => (
            <Skeleton key={index} className="h-12 w-full rounded-md" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-1 min-h-0 flex-col">
      <PageHeader className="justify-between gap-3">
        <div className="flex min-w-0 items-center gap-1.5">
          <WorkspaceAvatar name={workspace?.name ?? "W"} size="sm" />
          <span className="truncate text-sm text-muted-foreground">{workspace?.name ?? "Workspace"}</span>
          <ChevronRight className="size-3 shrink-0 text-muted-foreground" />
          <span className="text-sm font-medium">Leads</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setApolloOpen(true)}
            disabled={!apolloConfigured}
            title={apolloConfigured ? "Apollo" : "Apollo not configured"}
          >
            <Search />
            Apollo
          </Button>
          <Button variant="outline" size="sm" onClick={() => setImportOpen(true)}>
            <Upload />
            Import
          </Button>
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus />
            New lead
          </Button>
        </div>
      </PageHeader>

      <div className="flex h-auto shrink-0 flex-wrap items-center gap-2 border-b px-5 py-2">
        <button
          type="button"
          onClick={() => setStatusFilter("all")}
          className={cn(
            "inline-flex h-7 items-center gap-1.5 rounded-md border px-2 text-xs transition-colors",
            statusFilter === "all" ? "bg-accent text-foreground" : "text-muted-foreground hover:bg-accent/60",
          )}
        >
          All
          <span className="tabular-nums">{leads.length}</span>
        </button>
        {STATUS_ORDER.map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => setStatusFilter(status)}
            className={cn(
              "inline-flex h-7 items-center gap-1.5 rounded-md border px-2 text-xs transition-colors",
              statusFilter === status ? "bg-accent text-foreground" : "text-muted-foreground hover:bg-accent/60",
            )}
          >
            {statusFilter === status && <Check className="size-3" />}
            {STATUS_CONFIG[status].label}
          </button>
        ))}
        {hotCount > 0 && (
          <span className="ml-auto inline-flex h-7 items-center gap-1.5 rounded-md bg-red-50 px-2 text-xs font-medium text-red-700 dark:bg-red-950/40 dark:text-red-300">
            <Flame className="size-3" />
            {hotCount}
          </span>
        )}
      </div>

      {leads.length === 0 ? (
        <div className="flex flex-1 min-h-0 flex-col items-center justify-center gap-3 text-muted-foreground">
          <UsersRound className="size-10 text-muted-foreground/40" />
          <p className="text-sm">No leads yet</p>
          <div className="flex items-center gap-2">
            <Button size="sm" onClick={() => setCreateOpen(true)}>
              <Plus />
              New lead
            </Button>
            <Button size="sm" variant="outline" onClick={() => setApolloOpen(true)} disabled={!apolloConfigured}>
              <Search />
              Apollo
            </Button>
            <Button size="sm" variant="outline" onClick={() => setImportOpen(true)}>
              <Upload />
              Import
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex-1 min-h-0 overflow-auto">
          <LeadsTable leads={leads} />
        </div>
      )}

      <CreateLeadDialog open={createOpen} onOpenChange={setCreateOpen} />
      <ImportCsvDialog open={importOpen} onOpenChange={setImportOpen} />
      <ApolloImportDialog open={apolloOpen} onOpenChange={setApolloOpen} configured={apolloConfigured} />
    </div>
  );
}
