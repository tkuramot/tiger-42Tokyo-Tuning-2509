import {
  baseApiUrl,
  baseClientUrl,
  baseRobotUrl,
  baseLoginUrl,
} from "./config.js";

const createUrl = (baseUrl, path) => {
  const normalizedBaseUrl = baseUrl.endsWith("/")
    ? baseUrl.slice(0, -1)
    : baseUrl;
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${normalizedBaseUrl}${normalizedPath}`;
};

export const createClientUrl = (path) => createUrl(baseClientUrl, path);
export const createApiUrl = (path) => createUrl(baseApiUrl, path);
export const createRobotUrl = (path) => createUrl(baseRobotUrl, path);
export const createLoginUrl = (path) => createUrl(baseLoginUrl, path);
