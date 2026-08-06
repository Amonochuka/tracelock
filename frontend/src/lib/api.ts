const browserApiUrl = typeof window === 'undefined'
  ? 'http://localhost:8080'
  : `${window.location.protocol}//${window.location.hostname}:8080`;

export const API_URL = process.env.NEXT_PUBLIC_API_URL || browserApiUrl;

export const WS_URL = API_URL
  .replace("https://", "wss://")
  .replace("http://", "ws://");
