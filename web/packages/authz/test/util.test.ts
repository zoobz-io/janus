import { describe, expect, it } from "vitest";

import { loginUrl, settle } from "../src/util";

describe("loginUrl", () => {
  it("appends the hosted flow path to an absolute origin", () => {
    expect(loginUrl("https://janus.example")).toBe(
      "https://janus.example/auth/login",
    );
  });

  it("appends the flow path to a same-origin proxy prefix", () => {
    expect(loginUrl("/api/janus")).toBe("/api/janus/auth/login");
  });

  it("strips trailing slashes so the path never doubles up", () => {
    expect(loginUrl("https://janus.example/")).toBe(
      "https://janus.example/auth/login",
    );
    expect(loginUrl("/api/janus///")).toBe("/api/janus/auth/login");
  });
});

describe("settle", () => {
  it("returns the parsed JSON body for a 2xx response", async () => {
    const res = new Response(JSON.stringify({ id: "u_1" }), { status: 200 });

    await expect(settle<{ id: string }>(res)).resolves.toEqual({ id: "u_1" });
  });

  it("returns null for 401 (the caller isn't signed in)", async () => {
    const res = new Response(null, { status: 401 });

    await expect(settle(res)).resolves.toBeNull();
  });

  it("throws on any other non-ok status, naming the status", async () => {
    const res = new Response("boom", { status: 500 });

    await expect(settle(res)).rejects.toThrow(/500/);
  });
});
