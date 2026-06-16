import type { Citation } from "./types";

export function citationLabel(citation: Citation): string {
  const path = citation.path || "unknown";
  const start = citation.start_line || 0;
  const end = citation.end_line || start;
  const commit = citation.commit_sha ? ` @ ${citation.commit_sha.slice(0, 7)}` : "";
  return `${path}:${start}-${end}${commit}`;
}

export function currency(value?: number): string {
  if (!value) {
    return "$0.000000";
  }
  return `$${value.toFixed(6)}`;
}

export function compactList(values?: string[]): string {
  if (!values || values.length === 0) {
    return "none";
  }
  return values.join(" -> ");
}
