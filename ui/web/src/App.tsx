import { FormEvent, useMemo, useState } from "react";
import {
  AlertTriangle,
  CheckCircle2,
  ChevronRight,
  Clipboard,
  Code2,
  Database,
  Download,
  FileSearch,
  FileText,
  GitBranch,
  Loader2,
  LogIn,
  MessageSquareText,
  Play,
  Search,
  Send,
  ShieldCheck
} from "lucide-react";
import {
  apiBaseUrl,
  analyzeImpact,
  askRepository,
  createRepository,
  generateDeepDiveReport,
  generatePlan,
  getTrace,
  login,
  saveFeedback,
  startIngestion
} from "./api";
import { citationLabel, compactList, currency } from "./format";
import { arrayOrEmpty, confidenceLabel, formatEvidence, formatList, formatReportProvenance, formatReportQuality } from "./reportFormat";
import type {
  AskResponse,
  Citation,
  DeepDiveReportResponse,
  DeepDiveReportSection,
  FeedbackPayload,
  ImpactAnalysisResponse,
  ImplementationPlanResponse,
  IngestionJob,
  Repository
} from "./types";

type RerankerMode = "adaptive" | "on" | "off";

const storedToken = localStorage.getItem("kf_token") || "";
const storedEmail = localStorage.getItem("kf_email") || "";
const demoToken = new URLSearchParams(window.location.search).get("demo") === "1" ? "demo-token" : "";

export function App() {
  const [token, setToken] = useState(storedToken || demoToken);
  const [email, setEmail] = useState(storedEmail || (demoToken ? "demo@example.com" : "admin@example.com"));
  const [password, setPassword] = useState("");
  const [repository, setRepository] = useState<Repository | null>(null);
  const [job, setJob] = useState<IngestionJob | null>(null);
  const [question, setQuestion] = useState("Where is authentication implemented?");
  const [topK, setTopK] = useState(5);
  const [rerankerMode, setRerankerMode] = useState<RerankerMode>("adaptive");
  const [answer, setAnswer] = useState<AskResponse | null>(null);
  const [plan, setPlan] = useState<ImplementationPlanResponse | null>(null);
  const [impact, setImpact] = useState<ImpactAnalysisResponse | null>(null);
  const [report, setReport] = useState<DeepDiveReportResponse | null>(null);
  const [trace, setTrace] = useState<unknown>(null);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [reportCopied, setReportCopied] = useState(false);
  const [feedback, setFeedback] = useState<FeedbackPayload>({
    trace_id: "",
    answer_correct: true,
    citation_correct: true,
    missing_file: false,
    missing_symbol: false,
    hallucinated_claim: false,
    should_have_refused: false,
    reviewer_note: ""
  });
  const [feedbackSaved, setFeedbackSaved] = useState(false);

  const citedFiles = useMemo(() => {
    const files = new Set((answer?.citations || []).map((citation) => citation.path).filter(Boolean));
    return Array.from(files) as string[];
  }, [answer]);

  async function handleLogin(event: FormEvent) {
    event.preventDefault();
    setError("");
    setBusy("login");
    try {
      const payload = await login(email, password);
      localStorage.setItem("kf_token", payload.access_token);
      localStorage.setItem("kf_email", payload.user.email);
      setToken(payload.access_token);
      setEmail(payload.user.email);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy("");
    }
  }

  async function handleRepository(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setBusy("repository");
    const data = new FormData(event.currentTarget);
    try {
      const repo = await createRepository(token, {
        name: String(data.get("name") || ""),
        remote_url: String(data.get("remote_url") || ""),
        local_path: String(data.get("local_path") || ""),
        default_branch: String(data.get("default_branch") || "main")
      });
      setRepository(repo);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Repository import failed");
    } finally {
      setBusy("");
    }
  }

  async function handleIngestion(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!repository) {
      return;
    }
    setError("");
    setBusy("ingestion");
    const data = new FormData(event.currentTarget);
    try {
      const payload = await startIngestion(token, repository.id, {
        branch_name: String(data.get("branch_name") || repository.default_branch || "main"),
        commit_sha: String(data.get("commit_sha") || ""),
        process_now: data.get("process_now") === "on"
      });
      setJob(payload);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Indexing failed");
    } finally {
      setBusy("");
    }
  }

  async function handleAsk() {
    if (!repository || !question.trim()) {
      return;
    }
    setError("");
    setTrace(null);
    setPlan(null);
    setImpact(null);
    setReport(null);
    setReportCopied(false);
    setFeedbackSaved(false);
    setBusy("ask");
    try {
      const payload = await askRepository(token, {
        repository_id: repository.id,
        branch_name: repository.default_branch || "main",
        question,
        top_k: topK,
        ...(rerankerMode === "adaptive" ? {} : { reranker_enabled: rerankerMode === "on" })
      });
      setAnswer(payload);
      setFeedback((prev) => ({ ...prev, trace_id: payload.trace_id }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Question failed");
    } finally {
      setBusy("");
    }
  }

  async function handlePlan() {
    if (!repository || !question.trim()) {
      return;
    }
    setError("");
    setBusy("plan");
    try {
      setPlan(await generatePlan(token, {
        repository_id: repository.id,
        branch_name: repository.default_branch || "main",
        request: question,
        top_k: topK,
        ...(rerankerMode === "adaptive" ? {} : { reranker_enabled: rerankerMode === "on" })
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Plan generation failed");
    } finally {
      setBusy("");
    }
  }

  async function handleImpact() {
    if (!repository || !question.trim()) {
      return;
    }
    setError("");
    setBusy("impact");
    try {
      setImpact(await analyzeImpact(token, {
        repository_id: repository.id,
        branch_name: repository.default_branch || "main",
        request: question,
        top_k: topK,
        ...(rerankerMode === "adaptive" ? {} : { reranker_enabled: rerankerMode === "on" })
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Impact analysis failed");
    } finally {
      setBusy("");
    }
  }

  async function handleReport() {
    if (!repository) {
      return;
    }
    setError("");
    setReportCopied(false);
    setBusy("report");
    try {
      setReport(await generateDeepDiveReport(token, {
        repository_id: repository.id,
        branch_name: repository.default_branch || "main",
        request: question,
        top_k: topK,
        ...(rerankerMode === "adaptive" ? {} : { reranker_enabled: rerankerMode === "on" })
      }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Deep-dive report failed");
    } finally {
      setBusy("");
    }
  }

  async function handleCopyReport() {
    if (!report?.markdown) {
      return;
    }
    await navigator.clipboard.writeText(report.markdown);
    setReportCopied(true);
  }

  function handleDownloadReport() {
    if (!report?.markdown) {
      return;
    }
    const blob = new Blob([report.markdown], { type: "text/markdown;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "knowledge-forge-deep-dive.md";
    anchor.click();
    URL.revokeObjectURL(url);
  }

  async function handleTrace() {
    if (!answer?.trace_id) {
      return;
    }
    setError("");
    setBusy("trace");
    try {
      setTrace(await getTrace(token, answer.trace_id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Trace lookup failed");
    } finally {
      setBusy("");
    }
  }

  async function handleFeedback() {
    if (!answer?.trace_id) {
      return;
    }
    setError("");
    setBusy("feedback");
    try {
      await saveFeedback(token, { ...feedback, trace_id: answer.trace_id });
      setFeedbackSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Feedback save failed");
    } finally {
      setBusy("");
    }
  }

  if (!token) {
    return (
      <main className="login-shell">
        <form className="login-panel" onSubmit={handleLogin}>
          <div className="brand-row">
            <Code2 size={28} aria-hidden />
            <h1>Knowledge Forge</h1>
          </div>
          <label>
            Email
            <input value={email} onChange={(event) => setEmail(event.target.value)} />
          </label>
          <label>
            Password
            <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </label>
          {error && <p className="error">{error}</p>}
          <button className="primary" type="submit" disabled={busy === "login"}>
            {busy === "login" ? <Loader2 className="spin" size={16} aria-hidden /> : <LogIn size={16} aria-hidden />}
            Login
          </button>
        </form>
      </main>
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand-row">
          <Code2 size={24} aria-hidden />
          <h1>Knowledge Forge</h1>
        </div>
        <div className="topbar-meta">
          <span>{email}</span>
          <span>{apiBaseUrl}</span>
          <button className="ghost" onClick={() => {
            localStorage.removeItem("kf_token");
            setToken("");
          }}>
            Logout
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          <AlertTriangle size={18} aria-hidden />
          <span>{error}</span>
        </div>
      )}

      <section className="workspace">
        <aside className="rail">
          <section className="panel">
            <div className="panel-title">
              <Database size={18} aria-hidden />
              <h2>Repository</h2>
            </div>
            <form className="stack" onSubmit={handleRepository}>
              <label>
                Name
                <input name="name" defaultValue="knowledge-forge-demo" />
              </label>
              <label>
                Remote URL
                <input name="remote_url" placeholder="https://github.com/org/repo.git" />
              </label>
              <label>
                Local path
                <input name="local_path" placeholder="/absolute/path/to/repo" />
              </label>
              <label>
                Branch
                <input name="default_branch" defaultValue="main" />
              </label>
              <button className="primary" type="submit" disabled={busy === "repository"}>
                {busy === "repository" ? <Loader2 className="spin" size={16} aria-hidden /> : <GitBranch size={16} aria-hidden />}
                Save
              </button>
            </form>
            {repository && (
              <div className="fact-list">
                <span>ID</span>
                <code>{repository.id}</code>
                <span>Branch</span>
                <code>{repository.default_branch}</code>
                <span>Status</span>
                <code>{repository.status}</code>
              </div>
            )}
          </section>

          <section className="panel">
            <div className="panel-title">
              <Play size={18} aria-hidden />
              <h2>Index</h2>
            </div>
            <form className="stack" onSubmit={handleIngestion}>
              <label>
                Branch
                <input name="branch_name" defaultValue={repository?.default_branch || "main"} />
              </label>
              <label>
                Commit SHA
                <input name="commit_sha" placeholder="optional" />
              </label>
              <label className="check-row">
                <input name="process_now" type="checkbox" />
                Process now
              </label>
              <button className="primary" type="submit" disabled={!repository || busy === "ingestion"}>
                {busy === "ingestion" ? <Loader2 className="spin" size={16} aria-hidden /> : <Send size={16} aria-hidden />}
                Queue
              </button>
            </form>
            {job && (
              <div className="fact-list">
                <span>Job</span>
                <code>{job.id}</code>
                <span>Status</span>
                <code>{job.status}</code>
                <span>Snapshot</span>
                <code>{job.snapshot_id || "pending"}</code>
              </div>
            )}
          </section>
        </aside>

        <section className="demo-surface">
          <section className="question-band">
            <div>
              <div className="panel-title">
                <MessageSquareText size={18} aria-hidden />
                <h2>Question</h2>
              </div>
              <textarea value={question} onChange={(event) => setQuestion(event.target.value)} />
            </div>
            <div className="controls">
              <label>
                Top K
                <div className="segmented">
                  {[5, 8].map((value) => (
                    <button key={value} type="button" className={topK === value ? "active" : ""} onClick={() => setTopK(value)}>
                      {value}
                    </button>
                  ))}
                </div>
              </label>
              <label>
                Reranker
                <select value={rerankerMode} onChange={(event) => setRerankerMode(event.target.value as RerankerMode)}>
                  <option value="adaptive">Adaptive</option>
                  <option value="on">On</option>
                  <option value="off">Off</option>
                </select>
              </label>
              <button className="primary ask" onClick={handleAsk} disabled={!repository || busy === "ask"}>
                {busy === "ask" ? <Loader2 className="spin" size={16} aria-hidden /> : <Search size={16} aria-hidden />}
                Ask
              </button>
            </div>
          </section>

          <section className="answer-grid">
            <article className="answer-panel">
              <div className="panel-title">
                <ShieldCheck size={18} aria-hidden />
                <h2>Answer</h2>
              </div>
              <p className="answer-text">{answer?.answer || "Ask a repository question."}</p>
              {answer?.provenance && (
                <div className="provenance-strip">
                  <Metric label="Category" value={answer.provenance.query_category} />
                  <Metric label="Path" value={compactList(answer.provenance.retrieval_path)} />
                  <Metric label="Context" value={`${answer.provenance.context_token_count} tokens`} />
                  <Metric label="Cost" value={currency(answer.provenance.estimated_cost_usd)} />
                </div>
              )}
            </article>

            <article className="panel evidence-panel">
              <div className="panel-title">
                <FileSearch size={18} aria-hidden />
                <h2>Evidence</h2>
              </div>
              <CitationList citations={answer?.citations || []} />
            </article>
          </section>

          <section className="report-panel panel">
            <div className="panel-heading">
              <div className="panel-title">
                <FileText size={18} aria-hidden />
                <h2>Deep-Dive Report</h2>
              </div>
              <div className="button-row">
                {report && (
                  <>
                    <button className="secondary compact" onClick={handleCopyReport}>
                      <Clipboard size={14} aria-hidden />
                      {reportCopied ? "Copied" : "Copy Markdown"}
                    </button>
                    <button className="secondary compact" onClick={handleDownloadReport}>
                      <Download size={14} aria-hidden />
                      Download Markdown
                    </button>
                  </>
                )}
                <button className="secondary compact" onClick={handleReport} disabled={!repository || busy === "report"}>
                  {busy === "report" ? <Loader2 className="spin" size={14} aria-hidden /> : <Send size={14} aria-hidden />}
                  Generate
                </button>
              </div>
            </div>
            {report ? (
              <>
                <SectionRows
                  rows={[
                    ["Summary", report.summary],
                    ["Evidence Quality", formatReportQuality(report)],
                    ["Provenance", formatReportProvenance(report)]
                  ]}
                />
                <div className="report-sections">
                  {arrayOrEmpty(report.sections).map((section) => (
                    <ReportSectionView key={section.id} section={section} />
                  ))}
                </div>
              </>
            ) : (
              <p className="muted">No report generated.</p>
            )}
          </section>

          <section className="workflow-grid">
            <article className="panel">
              <div className="panel-heading">
                <h2>Plan</h2>
                <button className="secondary compact" onClick={handlePlan} disabled={!repository || busy === "plan"}>
                  {busy === "plan" ? <Loader2 className="spin" size={14} aria-hidden /> : <Send size={14} aria-hidden />}
                  Generate
                </button>
              </div>
              <SectionRows
                rows={[
                  ["Observed Evidence", plan ? formatEvidence(plan.observed_evidence) : citedFiles.length ? citedFiles.join("\n") : "none"],
                  ["Recommended Changes", plan ? formatList(plan.recommended_changes) : "not generated"],
                  ["Assumptions", plan ? formatList(plan.assumptions) : answer ? "limited to cited repository evidence" : "none"],
                  ["Missing Context", plan ? formatList(plan.missing_context) : "not generated"],
                  ["Risks", plan ? formatList(plan.risks) : "not generated"],
                  ["Tests", plan ? formatList(plan.tests) : citedFiles.some((file) => file.includes("_test")) ? citedFiles.filter((file) => file.includes("_test")).join("\n") : "not identified"],
                  ["Confidence", plan ? confidenceLabel(plan.confidence) : "not generated"]
                ]}
              />
            </article>
            <article className="panel">
              <div className="panel-heading">
                <h2>Impact</h2>
                <button className="secondary compact" onClick={handleImpact} disabled={!repository || busy === "impact"}>
                  {busy === "impact" ? <Loader2 className="spin" size={14} aria-hidden /> : <Send size={14} aria-hidden />}
                  Analyze
                </button>
              </div>
              <SectionRows
                rows={[
                  ["Observed Evidence", impact ? formatEvidence(impact.observed_evidence) : citedFiles.length ? citedFiles.join("\n") : "none"],
                  ["Impacted Files", impact ? formatList(impact.impacted_files) : citedFiles.length ? citedFiles.join("\n") : "none"],
                  ["Impacted Symbols", impact ? formatList(impact.impacted_symbols) : "not generated"],
                  ["Affected Tests", impact ? formatList(impact.affected_tests) : "not identified"],
                  ["Dependency Reasoning", impact ? formatList(impact.dependency_reasoning) : "not generated"],
                  ["Risk", impact ? impact.risk_level : answer?.provenance?.context_token_count ? "evidence available" : "unknown"],
                  ["Missing Context", impact ? formatList(impact.missing_context) : answer?.citations?.length ? "none flagged" : "no citations yet"],
                  ["Confidence", impact ? confidenceLabel(impact.confidence) : "not generated"]
                ]}
              />
            </article>
          </section>

          <section className="developer-section">
            <details>
              <summary>
                <ChevronRight size={16} aria-hidden />
                Developer Tools
              </summary>
              <div className="developer-grid">
                <div className="panel">
                  <h2>Trace</h2>
                  <button className="secondary" onClick={handleTrace} disabled={!answer?.trace_id || busy === "trace"}>
                    {busy === "trace" ? <Loader2 className="spin" size={16} aria-hidden /> : <Search size={16} aria-hidden />}
                    Load
                  </button>
                  <pre>{trace ? JSON.stringify(trace, null, 2) : answer?.trace_id || "none"}</pre>
                </div>
                <div className="panel">
                  <h2>Feedback</h2>
                  <FeedbackForm feedback={feedback} setFeedback={setFeedback} onSave={handleFeedback} disabled={!answer?.trace_id || busy === "feedback"} />
                  {feedbackSaved && (
                    <p className="saved">
                      <CheckCircle2 size={16} aria-hidden />
                      Saved
                    </p>
                  )}
                </div>
              </div>
            </details>
          </section>
        </section>
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function CitationList({ citations }: { citations: Citation[] }) {
  const normalized = arrayOrEmpty(citations);
  if (normalized.length === 0) {
    return <p className="muted">No citations.</p>;
  }
  return (
    <ol className="citations">
      {normalized.map((citation) => (
        <li key={citation.chunk_id}>
          <code>{citationLabel(citation)}</code>
          <p>{citation.excerpt}</p>
        </li>
      ))}
    </ol>
  );
}

function SectionRows({ rows }: { rows: Array<[string, string]> }) {
  return (
    <div className="section-rows">
      {rows.map(([label, value]) => (
        <div key={label}>
          <span>{label}</span>
          <pre>{value}</pre>
        </div>
      ))}
    </div>
  );
}

function ReportSectionView({ section }: { section: DeepDiveReportSection }) {
  return (
    <details className="report-section-row">
      <summary>
        <ChevronRight size={15} aria-hidden />
        <span>{section.title}</span>
        <strong>{section.confidence?.label || "Low"}</strong>
      </summary>
      <SectionRows
        rows={[
          ["Findings", formatList(section.findings)],
          ["Missing Context", formatList(section.missing_context)],
          ["Evidence", formatEvidence(section.evidence)],
          ["Targeted Retrieval", section.targeted ? "yes" : "shared evidence pass"]
        ]}
      />
    </details>
  );
}

function FeedbackForm({
  feedback,
  setFeedback,
  onSave,
  disabled
}: {
  feedback: FeedbackPayload;
  setFeedback: (feedback: FeedbackPayload) => void;
  onSave: () => void;
  disabled: boolean;
}) {
  const flags: Array<[keyof FeedbackPayload, string]> = [
    ["answer_correct", "Answer correct"],
    ["citation_correct", "Citation correct"],
    ["missing_file", "Missing file"],
    ["missing_symbol", "Missing symbol"],
    ["hallucinated_claim", "Hallucinated claim"],
    ["should_have_refused", "Should have refused"]
  ];
  return (
    <div className="feedback">
      {flags.map(([key, label]) => (
        <label className="check-row" key={key}>
          <input
            type="checkbox"
            checked={Boolean(feedback[key])}
            onChange={(event) => setFeedback({ ...feedback, [key]: event.target.checked })}
          />
          {label}
        </label>
      ))}
      <textarea
        value={feedback.reviewer_note}
        onChange={(event) => setFeedback({ ...feedback, reviewer_note: event.target.value })}
        placeholder="Reviewer note"
      />
      <button className="secondary" onClick={onSave} disabled={disabled}>
        Save
      </button>
    </div>
  );
}
