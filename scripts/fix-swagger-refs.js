#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

// Path to the unified API file
const UNIFIED_API_PATH = path.join(
  __dirname,
  "../docs/swagger/unified-api.json",
);

// Known problematic references and their corrections
const REF_FIXES = {
  // Auth service references
  authAnalyticsGetUserStatsResponse: "analyticsGetUserStatsResponse",
  authNotificationsServiceUpdateUserPreferencesBody:
    "NotificationsServiceUpdateUserPreferencesBody",

  // Server monitor event references
  serverServerMonitorEvent: "serverServerMonitorEvent",
  serverMonitorEvent: "serverServerMonitorEvent",
  serverServerServerMonitorEvent: "serverServerMonitorEvent",

  // Add more fixes as needed
};

function fixReferences(obj) {
  if (typeof obj !== "object" || obj === null) {
    return;
  }

  // Check if this is a $ref
  if (obj.$ref && typeof obj.$ref === "string") {
    const refParts = obj.$ref.split("/");
    const defName = refParts[refParts.length - 1];

    // Check if this reference needs fixing
    if (REF_FIXES[defName]) {
      console.log(`  Fixing reference: ${defName} -> ${REF_FIXES[defName]}`);
      obj.$ref = `#/definitions/${REF_FIXES[defName]}`;
    }
  }

  // Recursively process all properties
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      fixReferences(obj[key]);
    }
  }
}

function findAllReferences(obj, refs = new Set()) {
  if (typeof obj !== "object" || obj === null) {
    return refs;
  }

  // Check if this is a $ref
  if (obj.$ref && typeof obj.$ref === "string") {
    const refParts = obj.$ref.split("/");
    const defName = refParts[refParts.length - 1];
    refs.add(defName);
  }

  // Recursively process all properties
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      findAllReferences(obj[key], refs);
    }
  }

  return refs;
}

function main() {
  console.log("🔧 Fixing Swagger references...\n");

  try {
    // Read the unified API file
    const content = fs.readFileSync(UNIFIED_API_PATH, "utf8");
    const api = JSON.parse(content);

    // Find all references
    const allRefs = findAllReferences(api);
    const definedDefs = new Set(Object.keys(api.definitions || {}));

    // Find missing definitions
    const missingDefs = Array.from(allRefs).filter(
      (ref) => !definedDefs.has(ref),
    );

    if (missingDefs.length > 0) {
      console.log("❌ Missing definitions found:");
      missingDefs.forEach((def) => console.log(`  - ${def}`));
      console.log("");
    }

    // Fix known problematic references
    console.log("🔄 Applying reference fixes:");
    fixReferences(api);

    // Try to auto-fix remaining missing references
    console.log("\n🔍 Attempting auto-fixes for remaining issues:");
    const autoFixed = new Map();

    missingDefs.forEach((missingDef) => {
      // Skip if already in REF_FIXES
      if (REF_FIXES[missingDef]) return;

      // Try to find a similar definition
      const possibleMatches = Array.from(definedDefs).filter((def) => {
        const missingLower = missingDef.toLowerCase();
        const defLower = def.toLowerCase();

        // Check if the definition contains the key part of the missing reference
        return (
          defLower.includes(missingLower) ||
          missingLower.includes(defLower) ||
          // Check if removing service prefixes helps
          missingLower.replace(
            /^(auth|analytics|server|vpn|dpi|notifications)/,
            "",
          ) ===
            defLower.replace(
              /^(auth|analytics|server|vpn|dpi|notifications)/,
              "",
            )
        );
      });

      if (possibleMatches.length === 1) {
        console.log(`  Auto-fixing: ${missingDef} -> ${possibleMatches[0]}`);
        autoFixed.set(missingDef, possibleMatches[0]);
      } else if (possibleMatches.length > 1) {
        console.log(
          `  Multiple matches for ${missingDef}: ${possibleMatches.join(", ")}`,
        );

        // Try to pick the best match
        const bestMatch = possibleMatches.find((match) => {
          const missingBase = missingDef.replace(
            /^(auth|analytics|server|vpn|dpi|notifications)/,
            "",
          );
          const matchBase = match.replace(
            /^(auth|analytics|server|vpn|dpi|notifications)/,
            "",
          );
          return missingBase.toLowerCase() === matchBase.toLowerCase();
        });

        if (bestMatch) {
          console.log(`    Selected best match: ${bestMatch}`);
          autoFixed.set(missingDef, bestMatch);
        }
      }
    });

    // Apply auto-fixes
    if (autoFixed.size > 0) {
      console.log("\n🔄 Applying auto-fixes:");
      function applyAutoFixes(obj) {
        if (typeof obj !== "object" || obj === null) return;

        if (obj.$ref && typeof obj.$ref === "string") {
          const refParts = obj.$ref.split("/");
          const defName = refParts[refParts.length - 1];

          if (autoFixed.has(defName)) {
            console.log(
              `  Fixing reference: ${defName} -> ${autoFixed.get(defName)}`,
            );
            obj.$ref = `#/definitions/${autoFixed.get(defName)}`;
          }
        }

        for (const key in obj) {
          if (obj.hasOwnProperty(key)) {
            applyAutoFixes(obj[key]);
          }
        }
      }

      applyAutoFixes(api);
    }

    // Write the fixed file
    fs.writeFileSync(UNIFIED_API_PATH, JSON.stringify(api, null, 2));
    console.log("\n✅ Fixed unified API written to", UNIFIED_API_PATH);

    // Final check
    const finalRefs = findAllReferences(api);
    const finalMissing = Array.from(finalRefs).filter(
      (ref) => !definedDefs.has(ref),
    );

    if (finalMissing.length > 0) {
      console.log(
        "\n⚠️  Warning: Some references could not be fixed automatically:",
      );
      finalMissing.forEach((ref) => console.log(`  - ${ref}`));
      console.log(
        "\nYou may need to manually fix these or update the REF_FIXES mapping.",
      );
    } else {
      console.log("\n✨ All references successfully fixed!");
    }
  } catch (error) {
    console.error("❌ Error fixing swagger references:", error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { fixReferences };
