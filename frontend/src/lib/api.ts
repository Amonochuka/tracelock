export const API_URL = "https://tracelock-db.onrender.com";

export const WS_URL = API_URL
  .replace("https://", "wss://")
  .replace("http://", "ws://");