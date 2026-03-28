import { opsPilotStringify } from './opspilotStringify';

describe('opsPilotStringify', () => {
  describe('primitive and edge-case inputs', () => {
    it('should return "null" for null', () => {
      expect(opsPilotStringify(null)).toBe('null');
    });

    it('should return "undefined" for undefined', () => {
      expect(opsPilotStringify(undefined)).toBe('undefined');
    });

    it('should return string representation of numbers', () => {
      expect(opsPilotStringify(42 as any)).toBe('42');
    });

    it('should return the string itself', () => {
      expect(opsPilotStringify('hello' as any)).toBe('hello');
    });

    it('should return empty string for empty object', () => {
      expect(opsPilotStringify({})).toBe('');
    });
  });

  describe('simple key-value objects', () => {
    it('should format flat key-value pairs', () => {
      const result = opsPilotStringify({ name: 'Alice', age: 30 });
      expect(result).toBe('Name: Alice\nAge: 30');
    });

    it('should skip null and undefined values', () => {
      const result = opsPilotStringify({ name: 'Alice', missing: null, also: undefined });
      expect(result).toBe('Name: Alice');
    });

    it('should convert camelCase keys to spaced labels', () => {
      const result = opsPilotStringify({ operationName: 'GET /api', serviceName: 'frontend' });
      expect(result).toBe('Operation Name: GET /api\nService Name: frontend');
    });

    it('should capitalize the first letter of labels', () => {
      const result = opsPilotStringify({ spanID: '123' });
      expect(result).toBe('Span ID: 123');
    });
  });

  describe('nested objects as sections', () => {
    it('should format objects with only primitive values as sections', () => {
      const result = opsPilotStringify({
        process: {
          serviceName: 'my-service',
          version: '1.0',
        },
      });
      expect(result).toContain('Process');
      expect(result).toContain('Service Name: my-service');
      expect(result).toContain('Version: 1.0');
    });

    it('should add blank lines between sections', () => {
      const result = opsPilotStringify({
        section1: { a: 1 },
        section2: { b: 2 },
      });
      const lines = result.split('\n');
      // section1 header, a:1, blank, section2 header, b:2
      expect(lines).toEqual(['Section1', 'A: 1', '', 'Section2', 'B: 2']);
    });
  });

  describe('arrays', () => {
    it('should format simple arrays as comma-separated', () => {
      const result = opsPilotStringify({ items: [1, 2, 3] });
      expect(result).toBe('Items: 1, 2, 3');
    });

    it('should skip empty arrays', () => {
      const result = opsPilotStringify({ items: [], name: 'test' });
      expect(result).toBe('Name: test');
    });

    it('should format key-value pair objects in arrays', () => {
      const result = opsPilotStringify({
        tags: [
          { key: 'env', value: 'prod' },
          { key: 'region', value: 'us-east' },
        ],
      });
      expect(result).toContain('env: prod');
      expect(result).toContain('region: us-east');
    });

    it('should skip key-value pairs with empty values', () => {
      const result = opsPilotStringify({
        tags: [
          { key: 'env', value: 'prod' },
          { key: 'empty', value: '' },
          { key: 'nope', value: null },
        ],
      });
      expect(result).toContain('env: prod');
      expect(result).not.toContain('empty');
      expect(result).not.toContain('nope');
    });

    it('should format general objects in arrays with labels', () => {
      const result = opsPilotStringify({
        spans: [{ operationName: 'GET', duration: 100 }],
      });
      expect(result).toContain('Operation Name: GET');
      expect(result).toContain('Duration: 100');
    });
  });

  describe('deeply nested objects', () => {
    it('should recurse into objects that contain nested objects', () => {
      const result = opsPilotStringify({
        outer: {
          inner: {
            key1: 'val1',
            key2: 'val2',
          },
        },
      });
      // inner is a section (only primitives), outer is a wrapper (has nested objects)
      expect(result).toContain('Inner');
      expect(result).toContain('Key1: val1');
    });
  });

  describe('realistic trace data', () => {
    it('should format a span-like object', () => {
      const span = {
        traceID: 'abc123',
        spanID: 'def456',
        operationName: 'HTTP GET /api/users',
        duration: 1500,
        startTime: 1700000000,
        tags: [
          { key: 'http.method', value: 'GET' },
          { key: 'http.status_code', value: 200 },
        ],
        process: {
          serviceName: 'user-service',
          tags: [{ key: 'hostname', value: 'node-1' }],
        },
      };

      const result = opsPilotStringify(span);
      expect(result).toContain('Trace ID: abc123');
      expect(result).toContain('Span ID: def456');
      expect(result).toContain('Operation Name: HTTP GET /api/users');
      expect(result).toContain('Duration: 1500');
      expect(result).toContain('http.method: GET');
      expect(result).toContain('http.status_code: 200');
    });
  });
});
