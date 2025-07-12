#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

// Old to new service name mappings
const SERVICE_MAPPINGS = {
  AuthServiceService: "AuthService",
  AnalyticsServiceService: "AnalyticsService",
  VpnCoreServiceService: "VpnService",
  NotificationsServiceService: "NotificationsService",
  ServerManagerServiceService: "ServersService",
};

// Function to replace service names in a file
function updateFile(filePath) {
  try {
    let content = fs.readFileSync(filePath, "utf8");
    let modified = false;

    // Check if file contains any of the old service names
    for (const [oldName, newName] of Object.entries(SERVICE_MAPPINGS)) {
      if (content.includes(oldName)) {
        // Replace the old service name with the new one
        const regex = new RegExp(`\\b${oldName}\\b`, "g");
        content = content.replace(regex, newName);
        modified = true;
        console.log(`  Updating ${oldName} -> ${newName} in ${filePath}`);
      }
    }

    // Write back if modified
    if (modified) {
      fs.writeFileSync(filePath, content, "utf8");
      return true;
    }
    return false;
  } catch (error) {
    console.error(`Error processing ${filePath}:`, error.message);
    return false;
  }
}

// Function to find all TypeScript/JavaScript files
function findFiles(directory, results = []) {
  const entries = fs.readdirSync(directory, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(directory, entry.name);

    if (entry.isDirectory()) {
      // Skip certain directories
      if (
        ["node_modules", ".next", "dist", "build", "generated"].includes(
          entry.name,
        )
      ) {
        continue;
      }
      // Recursively search subdirectories
      findFiles(fullPath, results);
    } else if (entry.isFile()) {
      // Check if it's a TypeScript/JavaScript file
      if (/\.(ts|tsx|js|jsx)$/.test(entry.name)) {
        results.push(fullPath);
      }
    }
  }

  return results;
}

// Main function
function main() {
  const frontendDir = path.join(__dirname, "../frontend/src");

  console.log("🔄 Updating service names in TypeScript files...\n");
  console.log("Service mappings:");
  for (const [oldName, newName] of Object.entries(SERVICE_MAPPINGS)) {
    console.log(`  ${oldName} -> ${newName}`);
  }
  console.log("");

  const files = findFiles(frontendDir);
  console.log(`Found ${files.length} files to check\n`);

  let updatedCount = 0;
  for (const file of files) {
    if (updateFile(file)) {
      updatedCount++;
    }
  }

  console.log(`\n✅ Updated ${updatedCount} files`);
}

// Run the script
if (require.main === module) {
  main();
}

module.exports = { updateFile, SERVICE_MAPPINGS };
