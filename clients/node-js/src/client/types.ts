export interface FleareClientOptions {
  username?: string;
  password?: string;
  timeout?: number; // Timeout for requests
}

export interface CommandData {
  command: string;
  requestId?: string;
  key?: string;
  path?: string;
  body?: any;
}

export interface ServerResponse {
  requestId: string;
  data: any;
}
