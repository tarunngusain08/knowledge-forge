export type LoginResponse = {
  access_token: string;
  user: {
    id: string;
    email: string;
    role: string;
  };
};

export type Repository = {
  id: string;
  name: string;
  remote_url?: string;
  local_path?: string;
  default_branch: string;
  status: string;
};

export type IngestionJob = {
  id: string;
  repository_id: string;
  branch_name: string;
  commit_sha: string;
  status: string;
  attempts: number;
  error_message?: string;
  snapshot_id?: string;
};

export type Citation = {
  chunk_id: string;
  repository_id?: string;
  snapshot_id?: string;
  branch_name?: string;
  commit_sha?: string;
  path?: string;
  start_line?: number;
  end_line?: number;
  excerpt: string;
  dense_score?: number;
  lexical_rank?: number;
  fused_rank?: number;
  rerank_score?: number;
};

export type AskResponse = {
  answer: string;
  citations: Citation[];
  trace_id: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  provenance?: {
    repository_id: string;
    branch_name: string;
    snapshot_id: string;
    commit_sha: string;
    query_category: string;
    prompt_version: string;
    retrieval_config: Record<string, unknown>;
    retrieval_path: string[];
    retrieved_chunk_ids: string[];
    stage_contributions: Record<string, number>;
    context_token_count: number;
    estimated_cost_usd: number;
    model: string;
  };
};

export type EvidenceItem = {
  chunk_id: string;
  repository_id?: string;
  snapshot_id?: string;
  branch_name?: string;
  commit_sha?: string;
  path: string;
  start_line?: number;
  end_line?: number;
  excerpt: string;
  dense_score?: number;
  rerank_score?: number;
  reasons?: string[];
};

export type EvidenceConfidence = {
  label: "High" | "Medium" | "Low";
  score: number;
  evidence_coverage: number;
  citation_count: number;
  context_token_count: number;
  reasons: string[] | null;
};

export type ImplementationPlanResponse = {
  observed_evidence: EvidenceItem[];
  recommended_changes: string[];
  assumptions: string[];
  missing_context: string[];
  risks: string[];
  tests: string[];
  confidence: EvidenceConfidence;
  answer: string;
  citations: Citation[];
  trace_id: string;
  provenance: AskResponse["provenance"];
  model: string;
};

export type ImpactAnalysisResponse = {
  observed_evidence: EvidenceItem[];
  impacted_files: string[];
  impacted_symbols: string[];
  affected_tests: string[];
  dependency_reasoning: string[];
  risk_level: "High" | "Medium" | "Low";
  missing_context: string[];
  confidence: EvidenceConfidence;
  answer: string;
  citations: Citation[];
  trace_id: string;
  provenance: AskResponse["provenance"];
  model: string;
};

export type DeepDiveReportSection = {
  id: string;
  title: string;
  findings: string[] | null;
  missing_context: string[] | null;
  evidence: EvidenceItem[] | null;
  citations: Citation[] | null;
  confidence: EvidenceConfidence | null;
  trace_id?: string;
  targeted: boolean;
  answer?: string;
};

export type DeepDiveEvidenceQuality = {
  files_examined: number;
  cited_files: string[] | null;
  cited_symbols: string[] | null;
  citation_count: number;
  evidence_coverage: number;
  missing_context: string[] | null;
  confidence: EvidenceConfidence | null;
};

export type DeepDiveReportResponse = {
  summary: string;
  sections: DeepDiveReportSection[] | null;
  evidence_quality: DeepDiveEvidenceQuality | null;
  trace_ids: string[] | null;
  provenance: AskResponse["provenance"];
  model: string;
  generated_at: string;
  markdown: string;
};

export type FeedbackPayload = {
  trace_id: string;
  answer_correct: boolean;
  citation_correct: boolean;
  missing_file: boolean;
  missing_symbol: boolean;
  hallucinated_claim: boolean;
  should_have_refused: boolean;
  reviewer_note: string;
};
