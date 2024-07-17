import { Hono } from "hono";
import { handle } from "hono/aws-lambda";
import { bearerAuth } from "hono/bearer-auth";
import { z }  from "zod";

import {
  useHealthChecks,
  useHealthCheckResults,
} from "../library/database";

const API_KEY = process.env.API_KEY;

const {
  createHealthCheck,
  updateHealthCheck,
  deleteHealthCheck,
  getHealthCheck,
  listHealthChecks,
} = useHealthChecks();

const { deleteHealthCheckResults } = useHealthCheckResults();

const app = new Hono();

app.use("/*", bearerAuth({token: API_KEY}));

app.get("/health", async (ctx) => {
  return ctx.text("OK", 200);
});

app.post("/healthCheck", async (ctx) => {
  const Params = z.object({
    url: z.string(),
    method: z.enum(["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"]),
    headers: z.record(z.string(), z.string()).nullish().transform( x => x ?? undefined ),
    body: z.string().nullish().transform( x => x ?? undefined ),
    intervalSeconds: z.number(),
    expectedStatus: z.number(),
    expectedBody: z.string().nullish().transform( x => x ?? undefined ),
  });

  const body = await ctx.req.json();
  const { success: validParams, data: params } = Params.safeParse(body);
  
  if (!validParams) {
    return ctx.json({}, 400);
  }

  const healthCheck = await createHealthCheck(params);
  return ctx.json(healthCheck, 200);
});

app.put("/healthCheck/:healthCheckId", async (ctx) => {
  const Params = z.object({
    url: z.string(),
    method: z.enum(["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"]),
    headers: z.record(z.string(), z.string()).nullish().transform( x => x ?? undefined ),
    body: z.string().nullish().transform( x => x ?? undefined ),
    intervalSeconds: z.number(),
    expectedStatus: z.number(),
    expectedBody: z.string().nullish().transform( x => x ?? undefined ),
  });

  const healthCheckId = ctx.req.param("healthCheckId");

  const body = await ctx.req.json();
  const { success: validParams, data: params } = Params.safeParse(body);

  if (!validParams) {
    return ctx.json({}, 400);
  }
  
  const updatedHealthCheck = await updateHealthCheck({
    healthCheckId: healthCheckId,
    url: params.url,
    method: params.method,
    headers: params.headers,
    body: params.body,
    intervalSeconds: params.intervalSeconds,
    expectedStatus: params.expectedStatus,
    expectedBody: params.expectedBody,
  });

  return ctx.json(updatedHealthCheck, 200);
});

app.delete("/healthCheck/:healthCheckId", async (ctx) => {
  const healthCheckId = ctx.req.param("healthCheckId");

  await deleteHealthCheck({healthCheckId});

  return ctx.json({}, 200);
});

app.get("/healthCheck", async (ctx) => {
  const healthChecks = await listHealthChecks({});
  return ctx.json(healthChecks, 200);
});

app.get("/healthCheck/:healthCheckId", async (ctx) => {
  const healthCheckId = ctx.req.param("healthCheckId");
  
  const healthCheck = await getHealthCheck({healthCheckId});

  return ctx.json(healthCheck, 200);
});

app.post("/healthCheck/:healthCheckId/reset", async (ctx) => {
  const healthCheckId = ctx.req.param("healthCheckId");
  
  await deleteHealthCheckResults({healthCheckId});

  return ctx.json({}, 200);
});

export const handler = handle(app);