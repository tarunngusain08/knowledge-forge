import { describe, expect, it } from "vitest";
import { arrayOrEmpty, formatEvidence, formatList, formatReportProvenance, formatReportQuality } from "../src/reportFormat";
import type { DeepDiveReportResponse } from "../src/types";

describe("report formatting", () => {
  it("normalizes null arrays in deep-dive report payloads", () => {
    const report: DeepDiveReportResponse = {
      summary: "Report generated",
      sections: null,
      evidence_quality: {
        files_examined: 1,
        cited_files: null,
        cited_symbols: null,
        citation_count: 1,
        evidence_coverage: 0.7,
        missing_context: null,
        confidence: {
          label: "Medium",
          score: 0.67,
          evidence_coverage: 0.7,
          citation_count: 1,
          context_token_count: 800,
          reasons: null
        }
      },
      trace_ids: null,
      provenance: {
        repository_id: "repo-1",
        branch_name: "main",
        snapshot_id: "snapshot-1",
        commit_sha: "abcdef123456",
        query_category: "architecture_explanation",
        prompt_version: "v1",
        retrieval_config: {},
        retrieval_path: ["dense"],
        retrieved_chunk_ids: [],
        stage_contributions: {},
        context_token_count: 800,
        estimated_cost_usd: 0,
        model: "gemini"
      },
      model: "gemini",
      generated_at: "2026-06-17T00:00:00Z",
      markdown: "# Report"
    };

    expect(arrayOrEmpty(report.sections)).toEqual([]);
    expect(formatList(null)).toBe("none");
    expect(formatEvidence(null)).toBe("none");
    expect(formatReportQuality(report)).toContain("Cited symbols: none");
    expect(formatReportProvenance(report)).toContain("Traces: none");
  });
});

