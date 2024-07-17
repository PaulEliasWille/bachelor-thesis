import { DynamoDBDocument } from "@aws-sdk/lib-dynamodb";
import { DynamoDB } from "@aws-sdk/client-dynamodb";
import { nanoid } from "nanoid";

const HEALTH_CHECK_TABLE = process.env.HEALTH_CHECK_TABLE;
const HEALTH_CHECK_RESULT_TABLE = process.env.HEALTH_CHECK_RESULT_TABLE;
const HEALTH_CHECK_RESULT_TABLE_HEALTH_CHECK_ID_INDEX = process.env.HEALTH_CHECK_RESULT_TABLE_HEALTH_CHECK_ID_INDEX;

export const useHealthChecks = () => {
  const client = DynamoDBDocument.from(new DynamoDB());

  const createHealthCheck = async (params) => {
    const {
      url, method, headers, body,
      intervalSeconds, expectedStatus, expectedBody,
    } = params;

    const healthCheck = {
      healthCheckId: nanoid(),
      url,
      method,
      headers,
      body,
      intervalSeconds,
      expectedStatus,
      expectedBody,
    };

    const putParams = {
      TableName: HEALTH_CHECK_TABLE,
      Item: healthCheck,
    };

    await client.put(putParams);

    return healthCheck;
  };

  const updateHealthCheck = async (params) => {
    const {
      healthCheckId, url, method, headers,
      body, intervalSeconds,
      expectedStatus, expectedBody,
    } = params;

    const attributeNames = {
      "#url": "url",
      "#method": "method",
      "#headers": "headers",
      "#body": "body",
      "#intervalSeconds": "intervalSeconds",
      "#expectedStatus": "expectedStatus",
      "#expectedBody": "expectedBody",
    };

    const attributeValues = {
      ":url": url,
      ":method": method,
      ":intervalSeconds": intervalSeconds,
      ":expectedStatus": expectedStatus,
    };

    const setExpressions = [
      "#url=:url",
      "#method=:method",
      "#intervalSeconds=:intervalSeconds",
      "#expectedStatus=:expectedStatus"
    ];
    const removeExpressions = [];
    
    if (headers === undefined) {
      removeExpressions.push("#headers");
    } else {
      setExpressions.push("#headers=:headers");
      attributeValues["#headers"] = headers;
    }

    if (body === undefined) {
      removeExpressions.push("#body");
    } else {
      setExpressions.push("#body=:body");
      attributeValues["#body"] = body;
    }

    if (expectedBody === undefined) {
      removeExpressions.push("#expectedBody");
    } else {
      setExpressions.push("#expectedBody=:expectedBody");
      attributeValues["#expectedBody"] = expectedBody;
    }

    const setExpression = setExpressions.join(", ");
    const removeExpression = removeExpressions.join(", ");

    const updateExpression = removeExpressions.length > 0 ?
      `SET ${setExpression} REMOVE ${removeExpression}` :
      `SET ${setExpression}`;
    
    const updateParams = {
      TableName: HEALTH_CHECK_TABLE,
      Key: { healthCheckId },
      UpdateExpression: updateExpression,
      ExpressionAttributeNames: attributeNames,
      ExpressionAttributeValues: attributeValues,
    };

    await client.update(updateParams);

    return {
      healthCheckId: healthCheckId,
      url,
      method,
      headers,
      body,
      intervalSeconds,
      expectedStatus,
      expectedBody,
    };
  };

  const deleteHealthCheck = async (params) => {
    const { healthCheckId } = params;
    
    const deleteParams = {
      TableName: HEALTH_CHECK_TABLE,
      Key: {
        healthCheckId
      },
    };

    await client.delete(deleteParams);
  };

  const listHealthChecks = async (params) => {
    const scanParams = { TableName: HEALTH_CHECK_TABLE };
    const result = await client.scan(scanParams)
    return result.Items;
  };

  const getHealthCheck = async (params) => {
    const { healthCheckId } = params;
    
    const getParams = {
      TableName: HEALTH_CHECK_TABLE,
      Key: {
        healthCheckId,
      },
    };

    const result = await client.get(getParams);

    return result.Item;
  };

  return {
    createHealthCheck,
    updateHealthCheck,
    deleteHealthCheck,
    listHealthChecks,
    getHealthCheck,
  };
};

export const useHealthCheckResults = () => {
  const client = DynamoDBDocument.from(new DynamoDB());

  const createHealthCheckResult = async (params) => {
    const {
      healthCheck,
      timestamp, status, body,
    } = params;

    const healthCheckResult = {
      healthCheckResultId: nanoid(),
      healthCheckId: healthCheck.healthCheckId,
      healthCheck,
      timestamp,
      status,
      body,
    };

    const putParams = {
      TableName: HEALTH_CHECK_RESULT_TABLE,
      Item: healthCheckResult
    };

    await client.put(putParams);

    return healthCheckResult;
  };

  const deleteHealthCheckResults = async (params) => {
    const { healthCheckId } = params;

    const itemsToDelete = await listHealthCheckResults({healthCheckId});

    for (const item of itemsToDelete) {
      const deleteParams = {
        TableName: HEALTH_CHECK_RESULT_TABLE,
        Key: {
          healthCheckResultId: item.healthCheckResultId,
          timestamp: item.timestamp,
        },
      };

      await client.delete(deleteParams);
    }
  };

  const listHealthCheckResults = async (params) => {
    const { healthCheckId } = params;

    const queryParams = {
      TableName: HEALTH_CHECK_RESULT_TABLE,
      IndexName: HEALTH_CHECK_RESULT_TABLE_HEALTH_CHECK_ID_INDEX,
      KeyConditionExpression: "healthCheckId = :healthCheckId",
      ExpressionAttributeValues: {
        ":healthCheckId": healthCheckId,
      },
    };
  
    const healthCheckResults = [];
    do {
      const queryData = await client.query(queryParams);
      healthCheckResults.push(...queryData.Items);
      queryParams.ExclusiveStartKey = queryData.LastEvaluatedKey;
    } while (queryParams.ExclusiveStartKey);

    return healthCheckResults;
  };

  const getLastHealthCheckResult = async (params) => {
    const { healthCheckId } = params;

    const queryParams = {
      TableName: HEALTH_CHECK_RESULT_TABLE,
      IndexName: HEALTH_CHECK_RESULT_TABLE_HEALTH_CHECK_ID_INDEX,
      KeyConditionExpression: "healthCheckId = :healthCheckId",
      ExpressionAttributeValues: {
        ":healthCheckId": healthCheckId,
      },
      ScanIndexForward: false, 
      Limit: 1,
    };   
  
    const result = await client.query(queryParams);
    return result.Items.length === 1 ? result.Items[0] : undefined;
  };

  return {
    createHealthCheckResult,
    deleteHealthCheckResults,
    listHealthCheckResults,
    getLastHealthCheckResult,
  };
};