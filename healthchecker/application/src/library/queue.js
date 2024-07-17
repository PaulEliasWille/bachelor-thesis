import { SQS } from "@aws-sdk/client-sqs";

const WORKER_QUEUE_URL = process.env.WORKER_QUEUE_URL;

export const useHealthCheckWorkerQueue = () => {
  const client = new SQS();

  const enqueueHealthCheck = async (params) => {
    const { healthCheck } = params;

    await client.sendMessage({
      QueueUrl: WORKER_QUEUE_URL,
      MessageBody: JSON.stringify(healthCheck),
    });
  };

  return { enqueueHealthCheck };
};