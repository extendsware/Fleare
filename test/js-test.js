// load-test.js
const net = require("net");
const { performance } = require("perf_hooks");

const CONCURRENT_CONNECTIONS = 300;
const TEST_DURATION_MS = 1000;
let completedRequests = 0;
let failedRequests = 0;

function createClient(id) {
  return new Promise((resolve) => {
    const client = new net.Socket();
    const startTime = performance.now();

    client.connect(8080, "127.0.0.1", () => {
      const message = `Message ${id}`;
      client.write(message);
    });

    client.on("data", (data) => {
      const latency = performance.now() - startTime;
      completedRequests++;
      client.destroy();
      resolve(latency);
    });

    client.on("error", (err) => {
      console.log(err);

      failedRequests++;
      client.destroy();
      resolve(0);
    });

    setTimeout(() => {
      client.destroy();
      resolve(0);
    }, TEST_DURATION_MS);
  });
}

async function runTest() {
  console.log(
    `Starting load test with ${CONCURRENT_CONNECTIONS} concurrent connections...`,
  );
  const startTime = performance.now();

  const promises = [];
  for (let i = 0; i < CONCURRENT_CONNECTIONS; i++) {
    promises.push(createClient(i));
  }

  const latencies = await Promise.all(promises);
  const testDuration = performance.now() - startTime;

  // Calculate statistics
  const successfulLatencies = latencies.filter((l) => l > 0);
  const avgLatency =
    successfulLatencies.reduce((a, b) => a + b, 0) /
      successfulLatencies.length || 0;
  const minLatency = Math.min(...successfulLatencies) || 0;
  const maxLatency = Math.max(...successfulLatencies) || 0;

  console.log("\nTest Results:");
  console.log(`Duration:       ${testDuration.toFixed(2)}ms`);
  console.log(`Total Requests: ${completedRequests + failedRequests}`);
  console.log(`Successful:     ${completedRequests}`);
  console.log(`Failed:         ${failedRequests}`);
  console.log(
    `Requests/sec:   ${(completedRequests / (testDuration / 1000)).toFixed(2)}`,
  );
  console.log(`Avg Latency:    ${avgLatency.toFixed(2)}ms`);
  console.log(`Min Latency:    ${minLatency.toFixed(2)}ms`);
  console.log(`Max Latency:    ${maxLatency.toFixed(2)}ms`);
}

runTest();
