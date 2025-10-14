// Constants for formatting
const INDENT_SIZE = 2;
const CAMEL_CASE_REGEX = /([a-z])([A-Z])/g;

export function opsPilotStringify(obj: unknown): string {
  if (!obj || typeof obj !== 'object') {
    return String(obj);
  }

  const lines: string[] = [];
  let currentSection = '';

  function formatLabel(key: string): string {
    // Convert camelCase to "Camel Case" and capitalize first letter
    const spacedLabel = key.replace(CAMEL_CASE_REGEX, '$1 $2');
    return spacedLabel.charAt(0).toUpperCase() + spacedLabel.slice(1);
  }

  function formatValue(value: any, indent = 0): string {
    // Handle null/undefined
    if (value === null || value === undefined) {
      return '';
    }

    // Handle arrays
    if (Array.isArray(value)) {
      return formatArray(value, indent);
    }

    // Handle objects
    if (typeof value === 'object') {
      return formatObject(value, indent);
    }

    // Handle primitives
    return String(value);
  }

  function formatArray(value: any[], indent: number): string {
    if (value.length === 0) {
      return '';
    }

    // Simple array (no objects)
    const hasObjects = value.some((item) => typeof item === 'object' && item !== null);
    if (!hasObjects) {
      return value.join(', ');
    }

    // Array with objects
    const indentStr = ' '.repeat(INDENT_SIZE).repeat(indent + 1);
    const formattedItems = value
      .map((item) => {
        if (typeof item !== 'object' || item === null) {
          return `${indentStr}${String(item)}`;
        }

        const entries = Object.entries(item);

        // Special case: key-value pair object
        if (entries.length === 2 && item.key !== undefined && item.value !== undefined) {
          if (item.value === '' || item.value === null || item.value === undefined) {
            return '';
          }
          return `${indentStr}${item.key}: ${item.value}`;
        }

        // General object in array
        const itemLines: string[] = [];
        for (const [k, v] of entries) {
          const formattedValue = formatValue(v, indent + 1);
          if (formattedValue !== '') {
            itemLines.push(`${indentStr}${formatLabel(k)}: ${formattedValue}`);
          }
        }
        return itemLines.join('\n');
      })
      .filter((item) => item !== '');

    if (formattedItems.length === 0) {
      return '';
    }

    return '\n' + formattedItems.join('\n');
  }

  function formatObject(value: Record<string, any>, indent: number): string {
    const indentStr = ' '.repeat(INDENT_SIZE).repeat(indent + 1);
    const objectLines: string[] = [];

    for (const [k, v] of Object.entries(value)) {
      const formattedValue = formatValue(v, indent + 1);
      if (formattedValue !== '') {
        objectLines.push(`${indentStr}${formatLabel(k)}: ${formattedValue}`);
      }
    }

    if (objectLines.length === 0) {
      return '';
    }

    return '\n' + objectLines.join('\n');
  }

  function processObject(obj: any) {
    for (const [key, value] of Object.entries(obj)) {
      if (value && typeof value === 'object' && !Array.isArray(value)) {
        // Check if this object contains nested objects or only primitives
        const hasNestedObjects = Object.values(value).some((v) => v && typeof v === 'object');

        // Objects with only primitive values become section headers (e.g., userInfo: { name: "John", age: 30 })
        // This creates output like: "User Info\nName: John\nAge: 30"
        if (!hasNestedObjects && Object.keys(value).length > 0) {
          // It's a section header - output the section name and its primitive key-value pairs
          if (currentSection) {
            lines.push(''); // Add blank line between sections
          }
          currentSection = formatLabel(key);
          lines.push(currentSection);

          // Process the nested properties
          for (const [nestedKey, nestedValue] of Object.entries(value)) {
            const formattedValue = formatValue(nestedValue);
            if (formattedValue === '') {
              continue;
            }
            lines.push(`${formatLabel(nestedKey)}: ${formattedValue}`);
          }
        } else {
          // Object has nested objects - recurse to find sections deeper in the tree
          processObject(value);
        }
      } else {
        // Simple key-value pair
        const formattedValue = formatValue(value);
        if (formattedValue !== '') {
          lines.push(`${formatLabel(key)}: ${formattedValue}`);
        }
      }
    }
  }

  processObject(obj);
  return lines.join('\n');
}
