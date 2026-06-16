import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, request } from "../src/api";

describe("api client", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("omits content-type for empty GET requests and keeps auth headers", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ status: "ok" }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await request<{ status: string }>("/healthz", { token: "test-token" });

    expect(result).toEqual({ status: "ok" });
    const init = fetchMock.mock.calls[0][1] as RequestInit & { headers: Record<string, string> };
    expect(init.method).toBe("GET");
    expect(init.headers.Authorization).toBe("Bearer test-token");
    expect(init.headers.Accept).toBe("application/json");
    expect(init.headers["Content-Type"]).toBeUndefined();
  });

  it("parses structured API error envelopes", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ error: "invalid JSON body" }, { status: 400 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(request("/v1/ask", { method: "POST", body: { unknown: true } })).rejects.toMatchObject({
      name: "ApiError",
      message: "invalid JSON body",
      status: 400,
      path: "/v1/ask"
    } satisfies Partial<ApiError>);
  });

  it("returns undefined for empty success responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(request<void>("/empty")).resolves.toBeUndefined();
  });

  it("sets JSON content-type only when a body is sent", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ id: "repo-1" }));
    vi.stubGlobal("fetch", fetchMock);

    await request<{ id: string }>("/v1/repositories", {
      method: "POST",
      token: "test-token",
      body: { name: "demo" }
    });

    const init = fetchMock.mock.calls[0][1] as RequestInit & { headers: Record<string, string> };
    expect(init.headers["Content-Type"]).toBe("application/json");
    expect(init.body).toBe(JSON.stringify({ name: "demo" }));
  });
});

function jsonResponse(body: unknown, init: ResponseInit = {}) {
  return new Response(JSON.stringify(body), {
    status: init.status ?? 200,
    headers: {
      "Content-Type": "application/json",
      ...init.headers
    }
  });
}
