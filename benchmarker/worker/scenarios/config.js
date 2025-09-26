const origin = __ENV.CLIENT_ORIGIN_URL || "http://localhost";
export const baseClientUrl = origin;
export const baseLoginUrl = `${origin}/api`;
export const baseApiUrl = `${origin}/api/v1`;
export const baseRobotUrl = `${origin}/api/robot`;
