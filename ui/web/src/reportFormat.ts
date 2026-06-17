import type { DeepDiveReportResponse, EvidenceConfidence, EvidenceItem } from "./types";

export function arrayOrEmpty<T>(values?: T[] | null): T[] {
  return Array.isArray(values) ? values : [];
}

export function formatEvidence(items?: EvidenceItem[] | null): string {
  const normalized = arrayOrEmpty(items);
  if (normalized.length === 0) {
    return "none";
  }
  return normalized.map((item) => {
    const range = item.start_line && item.end_line ? `:${item.start_line}-${item.end_line}` : "";
    const commit = item.commit_sha ? ` @ ${item.commit_sha.slice(0, 12)}` : "";
    return `${item.path || "unknown"}${range}${commit}\n${item.excerpt || ""}`;
  }).join("\n\n");
}

export function formatList(values?: Array<string | null | undefined> | null): string {
  const normalized = arrayOrEmpty(values).map((value) => String(value || "").trim()).filter(Boolean);
  return normalized.length ? normalized.join("\n") : "none";
}

export function confidenceLabel(confidence?: EvidenceConfidence | null): string {
  if (!confidence) {
    return "Low (0%, evidence coverage 0%)\nconfidence metadata unavailable";
  }
  const percent = Math.round((confidence.score || 0) * 100);
  const coverage = Math.round((confidence.evidence_coverage || 0) * 100);
  return `${confidence.label || "Low"} (${percent}%, evidence coverage ${coverage}%)\n${formatList(confidence.reasons)}`;
}

export function formatReportQuality(report: DeepDiveReportResponse): string {
  const quality = report.evidence_quality;
  if (!quality) {
    return "none";
  }
  return [
    `Confidence: ${confidenceLabel(quality.confidence)}`,
    `Files examined: ${quality.files_examined || 0}`,
    `Citations: ${quality.citation_count || 0}`,
    `Evidence coverage: ${Math.round((quality.evidence_coverage || 0) * 100)}%`,
    `Cited files: ${formatList(quality.cited_files)}`,
    `Cited symbols: ${formatList(quality.cited_symbols)}`,
    `Missing context: ${formatList(quality.missing_context)}`
  ].join("\n");
}

export function formatReportProvenance(report: DeepDiveReportResponse): string {
  const provenance = report.provenance;
  if (!provenance) {
    return "none";
  }
  return [
    `Branch: ${provenance.branch_name || "unknown"}`,
    `Snapshot: ${provenance.snapshot_id || "unknown"}`,
    `Commit: ${provenance.commit_sha || "unknown"}`,
    `Model: ${report.model || "unknown"}`,
    `Generated: ${new Date(report.generated_at).toLocaleString()}`,
    `Traces: ${formatList(report.trace_ids)}`
  ].join("\n");
}

