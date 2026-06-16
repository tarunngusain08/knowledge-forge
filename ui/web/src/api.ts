import type { AskResponse, FeedbackPayload, IngestionJob, LoginResponse, Repository } from "./types";

const defaultBaseUrl = "http://localhost:8080";

export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || defaultBaseUrl;

type RequestOptions = {
  token?: string;
  body?: unknown;
  method?: string;
};

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    method: options.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {})
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body)
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed with status ${response.status}`);
  }
  return response.json() as Promise<T>;
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

export function getTrace(token: string, traceId: string): Promise<unknown> {
  return request<unknown>(`/v1/retrieval-traces/${traceId}`, { token });
}

export function saveFeedback(token: string, payload: FeedbackPayload): Promise<unknown> {
  return request<unknown>("/v1/feedback", { method: "POST", token, body: payload });
}
