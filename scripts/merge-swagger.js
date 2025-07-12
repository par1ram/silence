#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

// Configuration
const SWAGGER_DIR = path.join(__dirname, "../docs/swagger");
const OUTPUT_FILE = path.join(SWAGGER_DIR, "unified-api.json");

// Service mapping for better organization
const SERVICE_INFO = {
  auth: {
    name: "Auth",
    description: "Authentication and authorization services",
  },
  analytics: {
    name: "Analytics",
    description: "Analytics and metrics collection",
  },
  server: { name: "Servers", description: "VPN server management" },
  "server-manager": {
    name: "ServerManager",
    description: "VPN server management",
  },
  vpn: { name: "VPN", description: "VPN connection and configuration" },
  "vpn-core": { name: "VPNCore", description: "VPN core functionality" },
  dpi: { name: "DPI", description: "Deep Packet Inspection bypass" },
  "dpi-bypass": {
    name: "DPIBypass",
    description: "Deep Packet Inspection bypass",
  },
  notifications: {
    name: "Notifications",
    description: "Notification services",
  },
};

function readSwaggerFiles() {
  const files = fs.readdirSync(SWAGGER_DIR);
  const swaggerFiles = files.filter((file) => file.endsWith(".swagger.json"));

  const swaggerDocs = {};

  for (const file of swaggerFiles) {
    const filePath = path.join(SWAGGER_DIR, file);
    const content = fs.readFileSync(filePath, "utf8");
    const serviceName = file.replace(".swagger.json", "");

    try {
      swaggerDocs[serviceName] = JSON.parse(content);
    } catch (error) {
      console.error(`Error parsing ${file}:`, error.message);
    }
  }

  return swaggerDocs;
}

function mergeSwaggerDocs(swaggerDocs) {
  const unified = {
    swagger: "2.0",
    info: {
      title: "Silence VPN Platform API",
      version: "1.0.0",
      description:
        "Unified API for Silence VPN Platform - includes all services: Auth, Analytics, Server Manager, VPN Core, DPI Bypass, and Notifications",
      contact: {
        name: "Silence Team",
        email: "team@silence.com",
      },
      license: {
        name: "MIT",
        url: "https://opensource.org/licenses/MIT",
      },
    },
    host: "localhost:8080",
    basePath: "/api/v1",
    schemes: ["http", "https"],
    consumes: ["application/json"],
    produces: ["application/json"],
    securityDefinitions: {
      Bearer: {
        type: "apiKey",
        name: "Authorization",
        in: "header",
        description:
          'JWT Authorization header using the Bearer scheme. Example: "Authorization: Bearer {token}"',
      },
    },
    tags: [],
    paths: {},
    definitions: {},
  };

  // Create tags for each service
  Object.entries(SERVICE_INFO).forEach(([key, info]) => {
    unified.tags.push({
      name: info.name,
      description: info.description,
    });
  });

  // Add health tag
  unified.tags.push({
    name: "Health",
    description: "Health check endpoints",
  });

  // Merge paths and definitions
  Object.entries(swaggerDocs).forEach(([serviceName, doc]) => {
    if (!doc.paths || !doc.definitions) {
      console.warn(`Skipping ${serviceName}: missing paths or definitions`);
      return;
    }

    // Merge paths
    Object.entries(doc.paths).forEach(([path, methods]) => {
      Object.entries(methods).forEach(([method, operation]) => {
        // Add security to protected endpoints (skip health endpoints)
        if (!path.includes("/health") && !operation.security) {
          operation.security = [{ Bearer: [] }];
        }

        // Update tags for better organization
        const serviceInfo = SERVICE_INFO[serviceName];
        if (serviceInfo) {
          if (path.includes("/health")) {
            operation.tags = ["Health"];
          } else {
            operation.tags = [serviceInfo.name];
          }
        }

        // Add operation to unified paths
        if (!unified.paths[path]) {
          unified.paths[path] = {};
        }
        unified.paths[path][method] = operation;
      });
    });

    // Merge definitions with service prefix to avoid conflicts
    Object.entries(doc.definitions).forEach(([defName, definition]) => {
      // Skip common definitions that are already included
      if (defName.startsWith("protobuf") || defName.startsWith("rpc")) {
        if (!unified.definitions[defName]) {
          unified.definitions[defName] = definition;
        }
        return;
      }

      // Check if definition already has service prefix to avoid double prefixing
      const servicePrefix = serviceName.replace(/-/g, "");
      if (defName.toLowerCase().startsWith(servicePrefix.toLowerCase())) {
        // Already prefixed, just add as is
        unified.definitions[defName] = definition;
      } else {
        // Add service-prefixed definitions
        const prefixedName = `${servicePrefix}${defName.charAt(0).toUpperCase()}${defName.slice(1)}`;
        unified.definitions[prefixedName] = definition;
      }

      // Update references in the definition
      updateReferences(definition, serviceName);
    });
  });

  // Update all path references to use prefixed definitions
  Object.values(unified.paths).forEach((pathMethods) => {
    Object.values(pathMethods).forEach((operation) => {
      updatePathReferences(operation);
    });
  });

  return unified;
}

function updateReferences(obj, serviceName) {
  if (typeof obj !== "object" || obj === null) return;

  Object.keys(obj).forEach((key) => {
    if (key === "$ref" && typeof obj[key] === "string") {
      const refParts = obj[key].split("/");
      const defName = refParts[refParts.length - 1];

      // Don't prefix protobuf and rpc definitions
      if (defName.startsWith("protobuf") || defName.startsWith("rpc")) {
        return;
      }

      // Check if definition already has service prefix
      const servicePrefix = serviceName.replace(/-/g, "");
      if (defName.toLowerCase().startsWith(servicePrefix.toLowerCase())) {
        // Already prefixed, update ref as is
        obj[key] = `#/definitions/${defName}`;
      } else {
        // Update reference to use service prefix
        const prefixedName = `${servicePrefix}${defName.charAt(0).toUpperCase()}${defName.slice(1)}`;
        obj[key] = `#/definitions/${prefixedName}`;
      }
    } else if (typeof obj[key] === "object") {
      updateReferences(obj[key], serviceName);
    }
  });
}

function updatePathReferences(operation) {
  if (!operation) return;

  // Update response references
  if (operation.responses) {
    Object.values(operation.responses).forEach((response) => {
      if (response.schema && response.schema.$ref) {
        const refParts = response.schema.$ref.split("/");
        const defName = refParts[refParts.length - 1];

        // Try to find the correct service prefix
        const serviceName = detectServiceFromDefinition(defName);
        if (serviceName) {
          const servicePrefix = serviceName.replace(/-/g, "");
          if (defName.toLowerCase().startsWith(servicePrefix.toLowerCase())) {
            // Already prefixed
            response.schema.$ref = `#/definitions/${defName}`;
          } else {
            const prefixedName = `${servicePrefix}${defName.charAt(0).toUpperCase()}${defName.slice(1)}`;
            response.schema.$ref = `#/definitions/${prefixedName}`;
          }
        }
      }
    });
  }

  // Update parameter references
  if (operation.parameters) {
    operation.parameters.forEach((param) => {
      if (param.schema && param.schema.$ref) {
        const refParts = param.schema.$ref.split("/");
        const defName = refParts[refParts.length - 1];

        const serviceName = detectServiceFromDefinition(defName);
        if (serviceName) {
          const servicePrefix = serviceName.replace(/-/g, "");
          if (defName.toLowerCase().startsWith(servicePrefix.toLowerCase())) {
            // Already prefixed
            param.schema.$ref = `#/definitions/${defName}`;
          } else {
            const prefixedName = `${servicePrefix}${defName.charAt(0).toUpperCase()}${defName.slice(1)}`;
            param.schema.$ref = `#/definitions/${prefixedName}`;
          }
        }
      }
    });
  }
}

function detectServiceFromDefinition(defName) {
  // Simple heuristic to detect service from definition name
  if (
    defName.toLowerCase().includes("auth") ||
    defName.toLowerCase().includes("user")
  ) {
    return "auth";
  }
  if (
    defName.toLowerCase().includes("analytics") ||
    defName.toLowerCase().includes("metric")
  ) {
    return "analytics";
  }
  if (defName.toLowerCase().includes("server")) {
    // Check for server-manager specific definitions
    if (defName.toLowerCase().includes("manager")) {
      return "server-manager";
    }
    return "server";
  }
  if (
    defName.toLowerCase().includes("vpn") ||
    defName.toLowerCase().includes("tunnel") ||
    defName.toLowerCase().includes("peer")
  ) {
    return "vpn";
  }
  if (
    defName.toLowerCase().includes("dpi") ||
    defName.toLowerCase().includes("bypass")
  ) {
    return "dpi";
  }
  if (
    defName.toLowerCase().includes("notification") ||
    defName.toLowerCase().includes("template")
  ) {
    return "notifications";
  }

  return null;
}

function main() {
  console.log("🔄 Merging Swagger files...");

  try {
    const swaggerDocs = readSwaggerFiles();
    console.log(`📖 Found ${Object.keys(swaggerDocs).length} swagger files`);

    const unified = mergeSwaggerDocs(swaggerDocs);
    console.log(
      `🔗 Merged ${Object.keys(unified.paths).length} paths and ${Object.keys(unified.definitions).length} definitions`,
    );

    fs.writeFileSync(OUTPUT_FILE, JSON.stringify(unified, null, 2));
    console.log(`✅ Unified API written to ${OUTPUT_FILE}`);

    // Generate summary
    const pathCount = Object.keys(unified.paths).length;
    const defCount = Object.keys(unified.definitions).length;
    const tagCount = unified.tags.length;

    console.log("\n📊 Summary:");
    console.log(`  • ${pathCount} API endpoints`);
    console.log(`  • ${defCount} data definitions`);
    console.log(`  • ${tagCount} service tags`);
    console.log(
      `  • Services: ${Object.values(SERVICE_INFO)
        .map((s) => s.name)
        .join(", ")}`,
    );
  } catch (error) {
    console.error("❌ Error merging swagger files:", error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { mergeSwaggerDocs, readSwaggerFiles };
