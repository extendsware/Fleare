import net from "net";
import { FleareClientOptions, CommandData, ServerResponse } from "./types";
import { parseJson, parseString, uniqueString } from "../utils/utils";
import { FleareError, AuthenticationError } from "../errors/errors";

export class FleareClient {
  private host: string;
  private port: number;
  private username?: string;
  private password?: string;
  private client: net.Socket;
  private isAuthenticated: boolean;
  private pendingRequests: Map<
    string,
    { resolve: (value: any) => void; reject: (reason?: any) => void }
  >;
  private buffer: string;
  private timeout: number;

  constructor(host: string, port: number, options: FleareClientOptions = {}) {
    this.host = host;
    this.port = port;
    this.username = options.username;
    this.password = options.password;
    this.timeout = options.timeout || 5000; // Default timeout: 5 seconds
    this.client = new net.Socket();
    this.isAuthenticated = false;
    this.pendingRequests = new Map();
    this.buffer = "";

    this.client.on("data", (data) => this.handleData(data));
    this.client.on("error", (err) => this.handleError(err));
  }

  public connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.client.connect(this.port, this.host, () => {
        this.authenticate()
          .then(() => {
            this.client.emit("Connected");
            resolve();
          })
          .catch((err) => {
            reject(new AuthenticationError(err.message));
          });
      });

      this.client.once("error", (err) => {
        reject(new FleareError(`Connection failed: ${err.message}`));
      });
    });
  }

  private authenticate(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.username || !this.password) {
        reject(new AuthenticationError("Username and password are required"));
        return;
      }

      this.client.write(`${this.username}\n`);
      this.client.write(`${this.password}\n`);

      const timeoutId = setTimeout(() => {
        reject(new AuthenticationError("Authentication timed out"));
      }, this.timeout);

      this.client.once("data", (data) => {
        clearTimeout(timeoutId);
        const response = parseJson(data.toString().trim());
        if (response.data === "CONNECTED") {
          this.isAuthenticated = true;
          resolve();
        } else {
          reject(new AuthenticationError(response.data));
        }
      });
    });
  }

  private handleData(data: Buffer): void {
    this.buffer += data.toString();
    let delimiterIndex;
    while ((delimiterIndex = this.buffer.indexOf("\r\n")) !== -1) {
      const message = this.buffer.substring(0, delimiterIndex);
      this.buffer = this.buffer.substring(delimiterIndex + 2);
      this.processResponse(message);
    }
  }

  private processResponse(response: string): void {
    const msg: ServerResponse = parseJson(response);
    if (this.pendingRequests.has(msg.requestId)) {
      const { resolve, reject } = this.pendingRequests.get(msg.requestId)!;
      if (msg.data?.toString().startsWith("Error:")) {
        reject(new FleareError(msg.data));
      } else {
        resolve(msg.data);
      }
      this.pendingRequests.delete(msg.requestId);
    }
  }

  private handleError(err: Error): void {
    // this.client.emit("error", err);
  }

  public close(): void {
    this.client.end();
  }

  /**
   * Retrieve all keys from the server.
   * @returns A promise resolving to an array of keys.
   */
  public KEYS(): Promise<string[]> {
    const send: CommandData = {
      command: "KEYS",
    };
    return this.#commonWriter(send);
  }

  /**
   * Set a key-value pair on the server.
   * @param key The key to set.
   * @param data The value to associate with the key.
   * @returns A promise resolving when the operation is complete.
   */
  public SET(key: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "SET",
      key,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Retrieve the value associated with a key from the server.
   * @param key The key to retrieve.
   * @returns A promise resolving to the value associated with the key.
   */
  public GET(key: string): Promise<any> {
    const send: CommandData = {
      command: "GET",
      key,
    };
    return this.#commonWriter(send);
  }

  /**
   * Update an key in a object on the server.
   * @param key The key of the object.
   * @param path The path to the key item in the object.
   * @param data The new value for the key.
   * @returns A promise resolving when the operation is complete.
   */
  public UPDATE(key: string, path: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "UPDATE",
      key,
      path,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Add an item to a list on the server.
   * @param key The key of the list.
   * @param data The item to add to the list.
   * @returns A promise resolving when the operation is complete.
   */
  public LIST_ADD(key: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "LIST.ADD",
      key,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Retrieve an item from a list on the server.
   * @param key The key of the list.
   * @param path The path to the item in the list.
   * @returns A promise resolving to the item.
   */
  public LIST_GET(key: string, path: string): Promise<any> {
    const send: CommandData = {
      command: "LIST.GET",
      key,
      path,
    };
    return this.#commonWriter(send);
  }

  /**
   * Delete an item from a list on the server.
   * @param key The key of the list.
   * @param path The path to the item in the list.
   * @returns A promise resolving when the operation is complete.
   */
  public LIST_DELETE(key: string, path: string): Promise<void> {
    const send: CommandData = {
      command: "LIST.DELETE",
      key,
      path,
    };
    return this.#commonWriter(send);
  }

  /**
   * Update an item in a list on the server.
   * @param key The key of the list.
   * @param path The path to the item in the list.
   * @param data The new value for the item.
   * @returns A promise resolving when the operation is complete.
   */
  public LIST_UPDATE(key: string, path: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "LIST.UPDATE",
      key,
      path,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Set a value in a map on the server.
   * @param key The key of the map.
   * @param path The path to the value in the map.
   * @param data The value to set.
   * @returns A promise resolving when the operation is complete.
   */
  public MAP_SET(key: string, path: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "MAP.SET",
      key,
      path,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Retrieve a value from a map on the server.
   * @param key The key of the map.
   * @param path The path to the value in the map.
   * @returns A promise resolving to the value.
   */
  public MAP_GET(key: string, path: string): Promise<any> {
    const send: CommandData = {
      command: "MAP.GET",
      key,
      path,
    };
    return this.#commonWriter(send);
  }

  /**
   * Add a value to a map on the server.
   * @param key The key of the map.
   * @param path The path to the value in the map.
   * @param data The value to add.
   * @returns A promise resolving when the operation is complete.
   */
  public MAP_ADD(key: string, path: string, data: any): Promise<void> {
    const send: CommandData = {
      command: "MAP.ADD",
      key,
      path,
      body: data,
    };
    return this.#commonWriter(send);
  }

  /**
   * Delete a value from a map on the server.
   * @param key The key of the map.
   * @param path The path to the value in the map.
   * @returns A promise resolving when the operation is complete.
   */
  public MAP_DELETE(key: string, path: string): Promise<void> {
    const send: CommandData = {
      command: "MAP.DELETE",
      key,
      path,
    };
    return this.#commonWriter(send);
  }

  #commonWriter(commandData: CommandData): Promise<any> {
    if (!this.isAuthenticated) {
      throw new AuthenticationError("Client is not authenticated");
    }

    const requestId = uniqueString();
    commandData.requestId = requestId;
    const reqData = parseString(commandData);
    const fullMessage = `${reqData}\r\n`;

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        reject(new FleareError("Request timed out"));
        this.pendingRequests.delete(requestId);
      }, this.timeout);

      this.pendingRequests.set(requestId, {
        resolve: (value) => {
          clearTimeout(timeoutId);
          resolve(value);
        },
        reject: (reason) => {
          clearTimeout(timeoutId);
          reject(reason);
        },
      });

      this.client.write(fullMessage, (err) => {
        if (err) {
          reject(new FleareError(`Failed to send command: ${err.message}`));
          this.pendingRequests.delete(requestId);
        }
      });
    });
  }
}
