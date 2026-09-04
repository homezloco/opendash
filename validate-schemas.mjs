import { readFileSync } from "node:fs";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));

function jsonPointer(target, pointer) {
  if (pointer === "") return target;
  const parts = pointer.split("/").slice(1);
  let current = target;
  for (const part of parts) {
    const key = part.replace(/~1/g, "/").replace(/~0/g, "~");
    if (current && typeof current === "object" && key in current) {
      current = current[key];
    } else {
      throw new Error(`Could not resolve JSON pointer ${pointer} at ${part}`);
    }
  }
  return current;
}

function resolveSchema(ref, root, current) {
  if (ref.startsWith("#")) {
    return jsonPointer(root, decodeURIComponent(ref.slice(1)));
  }
  if (ref.startsWith("#/")) {
    return jsonPointer(root, ref.slice(1));
  }
  throw new Error(`Unsupported $ref: ${ref}`);
}

function typeOf(value) {
  if (value === null) return "null";
  if (Array.isArray(value)) return "array";
  return typeof value;
}

function validateType(value, type) {
  if (typeof type === "string") {
    if (type === "integer") {
      return typeof value === "number" && Number.isInteger(value);
    }
    return typeOf(value) === type;
  }
  if (Array.isArray(type)) {
    return type.some((t) => validateType(value, t));
  }
  return true;
}

function validateFormat(value, format) {
  if (typeof value !== "string") return true;
  if (format === "date-time") {
    return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:?\d{2})$/.test(value);
  }
  if (format === "uri") {
    return /^[a-z][a-z0-9+.-]*:/i.test(value);
  }
  if (format === "email") {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }
  if (format === "uuid") {
    return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
  }
  return true;
}

function validate(value, schema, root, path) {
  root = root || schema;
  const errors = [];

  if (schema === true) return errors;
  if (schema === false) {
    return [{ path, message: "value not allowed" }];
  }

  if (schema.$ref) {
    const resolved = resolveSchema(schema.$ref, root, schema);
    return validate(value, resolved, root, path);
  }

  if (schema.type && !validateType(value, schema.type)) {
    errors.push({ path, message: `expected type ${schema.type}, got ${typeOf(value)}` });
    return errors;
  }

  if (schema.enum && !schema.enum.some((e) => JSON.stringify(e) === JSON.stringify(value))) {
    errors.push({ path, message: `expected one of ${JSON.stringify(schema.enum)}` });
  }

  if (schema.const !== undefined && JSON.stringify(schema.const) !== JSON.stringify(value)) {
    errors.push({ path, message: `expected ${JSON.stringify(schema.const)}` });
  }

  if (typeof value === "string") {
    if (schema.minLength !== undefined && value.length < schema.minLength) {
      errors.push({ path, message: `string length ${value.length} < minLength ${schema.minLength}` });
    }
    if (schema.maxLength !== undefined && value.length > schema.maxLength) {
      errors.push({ path, message: `string length ${value.length} > maxLength ${schema.maxLength}` });
    }
    if (schema.pattern) {
      const re = new RegExp(schema.pattern);
      if (!re.test(value)) {
        errors.push({ path, message: `string does not match pattern ${schema.pattern}` });
      }
    }
    if (schema.format) {
      if (!validateFormat(value, schema.format)) {
        errors.push({ path, message: `string does not match format ${schema.format}` });
      }
    }
  }

  if (typeof value === "number") {
    if (schema.multipleOf !== undefined && value % schema.multipleOf !== 0) {
      errors.push({ path, message: `value not multiple of ${schema.multipleOf}` });
    }
    if (schema.minimum !== undefined && value < schema.minimum) {
      errors.push({ path, message: `value ${value} < minimum ${schema.minimum}` });
    }
    if (schema.maximum !== undefined && value > schema.maximum) {
      errors.push({ path, message: `value ${value} > maximum ${schema.maximum}` });
    }
    if (schema.exclusiveMinimum !== undefined && value <= schema.exclusiveMinimum) {
      errors.push({ path, message: `value ${value} <= exclusiveMinimum ${schema.exclusiveMinimum}` });
    }
    if (schema.exclusiveMaximum !== undefined && value >= schema.exclusiveMaximum) {
      errors.push({ path, message: `value ${value} >= exclusiveMaximum ${schema.exclusiveMaximum}` });
    }
  }

  if (Array.isArray(value)) {
    if (schema.minItems !== undefined && value.length < schema.minItems) {
      errors.push({ path, message: `array length ${value.length} < minItems ${schema.minItems}` });
    }
    if (schema.maxItems !== undefined && value.length > schema.maxItems) {
      errors.push({ path, message: `array length ${value.length} > maxItems ${schema.maxItems}` });
    }
    if (schema.uniqueItems) {
      const seen = new Set();
      for (const item of value) {
        const key = JSON.stringify(item);
        if (seen.has(key)) {
          errors.push({ path, message: "array contains duplicate items" });
          break;
        }
        seen.add(key);
      }
    }
    if (schema.items) {
      value.forEach((item, i) => {
        errors.push(...validate(item, schema.items, root, `${path}[${i}]`));
      });
    }
  }

  if (typeof value === "object" && value !== null && !Array.isArray(value)) {
    if (schema.required) {
      for (const key of schema.required) {
        if (!(key in value)) {
          errors.push({ path, message: `missing required property ${key}` });
        }
      }
    }

    if (schema.properties) {
      for (const [key, sub] of Object.entries(schema.properties)) {
        if (key in value) {
          errors.push(...validate(value[key], sub, root, `${path}.${key}`));
        }
      }
    }

    if (schema.additionalProperties === false) {
      for (const key of Object.keys(value)) {
        if (!schema.properties || !(key in schema.properties)) {
          errors.push({ path, message: `additional property ${key} not allowed` });
        }
      }
    } else if (typeof schema.additionalProperties === "object") {
      for (const [key, val] of Object.entries(value)) {
        if (!schema.properties || !(key in schema.properties)) {
          errors.push(...validate(val, schema.additionalProperties, root, `${path}.${key}`));
        }
      }
    }

    if (schema.patternProperties) {
      for (const [pattern, sub] of Object.entries(schema.patternProperties)) {
        const re = new RegExp(pattern);
        for (const [key, val] of Object.entries(value)) {
          if (re.test(key)) {
            errors.push(...validate(val, sub, root, `${path}.${key}`));
          }
        }
      }
    }
  }

  if (schema.allOf) {
    for (const sub of schema.allOf) {
      errors.push(...validate(value, sub, root, path));
    }
  }

  if (schema.anyOf) {
    const anyValid = schema.anyOf.some((sub) => validate(value, sub, root, path).length === 0);
    if (!anyValid) {
      errors.push({ path, message: "value does not match anyOf" });
    }
  }

  if (schema.oneOf) {
    const validCount = schema.oneOf.filter((sub) => validate(value, sub, root, path).length === 0).length;
    if (validCount !== 1) {
      errors.push({ path, message: `value matches ${validCount} oneOf schemas, expected 1` });
    }
  }

  return errors;
}

function loadJson(path) {
  const full = resolve(__dirname, path);
  const text = readFileSync(full, "utf8");
  return JSON.parse(text);
}

function main() {
  const schemas = {
    appManifest: loadJson("schemas/app-manifest.schema.json"),
    catalogIndex: loadJson("schemas/catalog-index.schema.json"),
  };

  const fixtures = [
    { name: "catalog index", schema: schemas.catalogIndex, path: "catalog/index.json" },
    { name: "uptime-kuma app manifest", schema: schemas.appManifest, path: "catalog/apps/uptime-kuma/app.json" },
  ];

  let failed = false;

  for (const fixture of fixtures) {
    const instance = loadJson(fixture.path);
    const errors = validate(instance, fixture.schema, fixture.schema, "$");
    if (errors.length === 0) {
      console.log(`✓ ${fixture.name} is valid`);
    } else {
      failed = true;
      console.error(`✗ ${fixture.name} failed validation:`);
      for (const err of errors) {
        console.error(`  ${err.path}: ${err.message}`);
      }
    }
  }

  if (failed) {
    process.exit(1);
  }

  console.log("All schema fixtures are valid.");
}

main();
