export function parseJson(jsonString: string): any {
    try {
      return JSON.parse(jsonString);
    } catch (err) {
      throw new Error(`Failed to parse JSON: ${jsonString}`);
    }
  }
  
  export function parseString(data: any): string {
    return JSON.stringify(data);
  }
  
  export function uniqueString(): string {
    return Math.random().toString(36).substring(2, 15);
  }