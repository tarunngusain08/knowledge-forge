import { describe, expect, it } from "vitest";
import { citationLabel, compactList, currency } from "../src/format";

describe("format helpers", () => {
  it("formats repository citation labels with short commit", () => {
    expect(citationLabel({
      chunk_id: "chunk-1",
      path: "internal/auth/service.go",
      start_line: 10,
      end_line: 18,
      commit_sha: "abcdef123456",
      excerpt: "AuthService owns login"
    })).toBe("internal/auth/service.go:10-18 @ abcdef1");
  });

  it("formats zero costs and retrieval paths", () => {
    expect(currency(0)).toBe("$0.000000");
    expect(compactList(["dense", "rerank"])).toBe("dense -> rerank");
  });
});
