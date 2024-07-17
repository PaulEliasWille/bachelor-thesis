import moment from "moment";

const DISPATCH_INTERVAL_SECONDS = +process.env.DISPATCH_INTERVAL_SECONDS;

import {
  useHealthChecks,
  useHealthCheckResults,
} from "../library/database";

import {
  useHealthCheckWorkerQueue,
} from "../library/queue";

const { listHealthChecks } = useHealthChecks();
const { getLastHealthCheckResult } = useHealthCheckResults();
const { enqueueHealthCheck } = useHealthCheckWorkerQueue();

export const handler = async (event, context) => {
  const healthChecks = await listHealthChecks({});
  for (const healthCheck of healthChecks) {
    try {
      const lastHealthCheckResult = await getLastHealthCheckResult({healthCheckId: healthCheck.healthCheckId});
      if (lastHealthCheckResult !== undefined) {
        const timeSinceLastHealthCheck = moment().diff(moment(lastHealthCheckResult.timestamp), "seconds");
        const timeTillNextHealthCheck = healthCheck.intervalSeconds - timeSinceLastHealthCheck;
        if (timeTillNextHealthCheck > DISPATCH_INTERVAL_SECONDS/2) continue;
      }
  
      await enqueueHealthCheck({healthCheck});
    } catch (err) {
      console.log("health check dispatch failed");
    }
  }

  return {};
}