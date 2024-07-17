import { Hono } from "hono";
import { handle } from "hono/aws-lambda";
import React from "react";
import { renderToString } from "react-dom/server";
import moment from "moment";

import {
  useHealthChecks,
  useHealthCheckResults,
} from "../library/database";

const { getHealthCheck } = useHealthChecks();
const { listHealthCheckResults } = useHealthCheckResults();

const app = new Hono();

app.get("/dashboard/:healthCheckId", async (ctx) => {
  const healthCheckId = ctx.req.param("healthCheckId");

  const healthCheck = await getHealthCheck({healthCheckId});
  const healthCheckResults = await listHealthCheckResults({healthCheckId});

  const series = healthCheckResults
    .map((healthCheckResult) => {
      const statusOK = healthCheckResult.healthCheck.expectedStatus === healthCheckResult.status;
      const bodyOK = healthCheckResult.healthCheck.expectedBody === undefined ||
        healthCheckResult.healthCheck.expectedBody === healthCheckResult.body;

      const OK = statusOK && bodyOK;

      return [moment(healthCheckResult.timestamp), OK];
    })
    .toSorted(([leftTimestamp, leftOK], [rightTimestamp, rightOK]) => rightTimestamp.diff(leftTimestamp, "seconds"));
  
  if (series.length === 0) {
    return ctx.html(renderToString(
      <html>
        <head></head>
        <body>
          <h1>{healthCheck.url}</h1>
          <ul>
            <li>Uptime: No Data</li>
            <li>Last Status: No Data</li>
            <li>Last Timestamp: No Data</li>
          </ul>
        </body>
      </html>
    ), 200);
  }

  const successSeries = series.filter(([_, OK]) => OK);

  const successRate = (successSeries.length / series.length * 100).toFixed(0);
  const mostRecentResult = series[0];

  const mostRecentResultTimestamp = mostRecentResult[0];
  const mostRecentResultOK = mostRecentResult[1];

  return ctx.html(renderToString(
    <html>
      <head></head>
      <body>
        <h1>{healthCheck.url}</h1>
        <h2>General</h2>
        <ul>
          <li>Uptime: {successRate}%</li>
          <li>Last Status: {mostRecentResultOK ? "Success" : "Error"}</li>
          <li>Last Timestamp: {mostRecentResultTimestamp.toString()}</li>
        </ul>
        <h2>Details</h2>
        <ul>
          {series.map(([timestamp, OK], index) => (
            <li key={index}>{timestamp.toString()}: {OK ? "Success" : "Error"}</li>
          ))}
        </ul>
      </body>
    </html>
  ), 200);
});

export const handler = handle(app);