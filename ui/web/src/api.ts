import type {
  AskResponse,
  FeedbackPayload,
  ImpactAnalysisResponse,
  ImplementationPlanResponse,
  IngestionJob,
  LoginResponse,
  Repository
} from "./types";

const defaultBaseUrl = "http://localhost:8080";

export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || defaultBaseUrl;

type RequestOptions = {
  token?: string;
  body?: unknown;
  method?: string;
};

export class ApiError extends Error {
  status: number;
  path: string;

  constructor(message: string, status: number, path: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.path = path;
  }
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    Accept: "application/json",
    ...(options.token ? { Authorization: `Bearer ${options.token}` } : {})
  };
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  const response = await fetch(`${apiBaseUrl}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body === undefined ? undefined : JSON.stringify(options.body)
  });
  if (!response.ok) {
    throw new ApiError(await errorMessage(response), response.status, path);
  }
  if (response.status === 204 || response.headers.get("content-length") === "0") {
    return undefined as T;
  }
  const text = await response.text();
  if (!text.trim()) {
    return undefined as T;
  }
  return JSON.parse(text) as T;
}

async function errorMessage(response: Response): Promise<string> {
  const fallback = `Request failed with status ${response.status}`;
  const contentType = response.headers.get("content-type") || "";
  const text = await response.text();
  if (!text.trim()) {
    return fallback;
  }
  if (contentType.includes("application/json")) {
    try {
      const payload = JSON.parse(text) as { error?: unknown; message?: unknown };
      if (typeof payload.error === "string" && payload.error.trim()) {
        return payload.error;
      }
      if (typeof payload.message === "string" && payload.message.trim()) {
        return payload.message;
      }
    } catch {
      return fallback;
    }
  }
  return text;
}

export function login(email: string, password: string): Promise<LoginResponse> {
  return request<LoginResponse>("/auth/login", { method: "POST", body: { email, password } });
}

export function createRepository(token: string, input: {
  name: string;
  remote_url: string;
  local_path: string;
  default_branch: string;
}): Promise<Repository> {
  return request<Repository>("/v1/repositories", { method: "POST", token, body: input });
}

export function startIngestion(token: string, repositoryId: string, input: {
  branch_name: string;
  commit_sha: string;
  process_now: boolean;
}): Promise<IngestionJob> {
  return request<IngestionJob>(`/v1/repositories/${repositoryId}/ingestions`, { method: "POST", token, body: input });
}

export function askRepository(token: string, input: {
  repository_id: string;
  branch_name: string;
  question: string;
  top_k: number;
  reranker_enabled?: boolean;
}): Promise<AskResponse> {
  return request<AskResponse>("/v1/ask", { method: "POST", token, body: input });
}

export function generatePlan(token: string, input: {
  repository_id: string;
  branch_name: string;
  request: string;
  top_k: number;
  reranker_enabled?: boolean;
}): Promise<ImplementationPlanResponse> {
  return request<ImplementationPlanResponse>("/v1/plans", { method: "POST", token, body: input });
}

export function analyzeImpact(token: string, input: {
  repository_id: string;
  branch_name: string;
  request: string;
  top_k: number;
  reranker_enabled?: boolean;
}): Promise<ImpactAnalysisResponse> {
  return request<ImpactAnalysisResponse>("/v1/impact", { method: "POST", token, body: input });
}

export function getTrace(token: string, traceId: string): Promise<unknown> {
  return request<unknown>(`/v1/retrieval-traces/${traceId}`, { token });
}

export function saveFeedback(token: string, payload: FeedbackPayload): Promise<unknown> {
  return request<unknown>("/v1/feedback", { method: "POST", token, body: payload });
}
