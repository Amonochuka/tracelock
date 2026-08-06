export const API_URL = "https://tracelock.onrender.com";

export const WS_URL = API_URL
  .replace("https://", "wss://")
  .replace("http://", "ws://");