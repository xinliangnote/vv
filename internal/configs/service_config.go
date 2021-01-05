package configs

// ServiceConfig the default service config
const ServiceConfig = `
{
  "clientLanguage": [
    "go"
  ],
  "percentage": 100,
  "serviceConfig": {
    "loadBalancingPolicy": "round_robin",
    "methodConfig": [
      {
        "name": [
          {
            "service": "",
            "method": ""
          }
        ],
        "waitForReady": true,
        "timeout": "2s",
        "retryPolicy": {
          "maxAttempts": 3,
          "initialBackoff": ".01s",
          "maxBackoff": ".01s",
          "backoffMultiplier": 1,
          "retryableStatusCodes": [
            "UNAVAILABLE",
            "INTERNAL",
            "ABORTED"
          ]
        }
      }
    ],
    "retryThrottling": {
      "maxTokens": 10,
      "tokenRatio": 0.1
    }
  }
}
`
