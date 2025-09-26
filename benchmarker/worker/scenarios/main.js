import { userJourneyScenario, userJourneyScenario2 } from "./userJourney.js";
import { robotAPIScenario } from "./robot.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";

export const options = {
  scenarios: {
    userJourneyScenario: __ENV.CLIENT_ORIGIN_URL.includes("http://tuning-nginx")
      ? {
          exec: "userJourneyScenario",
          executor: "shared-iterations",
          vus: 1,
          iterations: 100000,
          startTime: "0s",
          maxDuration: "30s",
        }
      : {
          exec: "userJourneyScenario",
          executor: "ramping-vus",
          startVUs: 1,
          stages: [
            { duration: "120s", target: 1 }, //
            { duration: "30s", target: 100 }, // 60秒で100vusにランプアップ
            { duration: "30s", target: 100 }, // 60秒間100vusを維持
          ],
          startTime: "0s",
          gracefulStop: "60s",
        },
    userJourneyScenario2: __ENV.CLIENT_ORIGIN_URL.includes(
      "http://tuning-nginx"
    )
      ? {
          exec: "userJourneyScenario2",
          executor: "shared-iterations",
          vus: 1,
          iterations: 100000,
          startTime: "0s",
          maxDuration: "30s",
        }
      : {
          exec: "userJourneyScenario2",
          executor: "ramping-vus",
          startVUs: 1,
          stages: [
            { duration: "120s", target: 1 }, //
            { duration: "60s", target: 100 }, // 60秒で100vusにランプアップ
            { duration: "60s", target: 100 }, // 60秒間100vusを維持
          ],
          startTime: "0s",
          gracefulStop: "60s",
        },
    robotAPIScenario: {
      exec: "robotAPIScenario",
      executor: "shared-iterations",
      vus: 1,
      iterations: 100000,
      startTime: "0s",
      maxDuration: __ENV.CLIENT_ORIGIN_URL.includes("http://tuning-nginx")
        ? "30s"
        : "240s",
    },
  },
};

export { userJourneyScenario, userJourneyScenario2, robotAPIScenario };

export function handleSummary(data) {
  const rawDataFilePath = __ENV.RAW_DATA_FILE_PATH;

  return {
    [rawDataFilePath]: JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: "→", enableColors: true }),
  };
}
