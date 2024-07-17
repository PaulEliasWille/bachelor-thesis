import moment from "moment";

import {
  useHealthChecks,
  useHealthCheckResults,
} from "../library/database";

import {
  useHealthCheckWorkerQueue,
} from "../library/queue";

const { createHealthCheckResult } = useHealthCheckResults();

export const handler = async (event, context) => {
  for (const message of event.Records) {
    try {
      const healthCheck = JSON.parse(message.body);

      let responseStatus = -1;
      let responseBody = undefined;
      let responseTimestamp = moment().toISOString();

      try {
        const request = new Request(
          healthCheck.url,
          {
            method: healthCheck.method,
            headers: healthCheck.headers,
            body: healthCheck.body,
            signal: AbortSignal.timeout(5000)
          },
        );
      
        const response = await fetch(request);

        responseStatus = response.status;
        responseBody = await response.text();
      } catch (err) {}

      await createHealthCheckResult({
        healthCheck,
        timestamp: responseTimestamp,
        status: responseStatus,
        body: responseBody,
      });
    } catch(err) {
      console.log(err);
      console.log("health check failed");
    }
  }

  return {};
}