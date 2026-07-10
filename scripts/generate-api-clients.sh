#!/usr/bin/env bash
# Generates typed API clients for iOS (Swift) and Android (Kotlin) from the OpenAPI spec.
# Requires: openapi-generator-cli (https://openapi-generator.tech/)
#
# Install:
#   brew install openapi-generator   # macOS
#   npm install -g @openapitools/openapi-generator-cli  # cross-platform

set -euo pipefail

SPEC="$(dirname "$0")/../backend/openapi.yaml"
IOS_OUT="$(dirname "$0")/../apps/ios/PlexusClient/Network/Generated"
ANDROID_OUT="$(dirname "$0")/../apps/android/app/src/main/java/com/plexus/client/api/generated"

echo "→ Generating Swift 5 client..."
openapi-generator generate \
  -i "$SPEC" \
  -g swift5 \
  -o "$IOS_OUT" \
  --additional-properties=projectName=PlexusAPI,podSummary="Plexus API Client",responseAs=AsyncAwait

echo "→ Generating Kotlin client..."
openapi-generator generate \
  -i "$SPEC" \
  -g kotlin \
  -o "$ANDROID_OUT" \
  --additional-properties=packageName=com.plexus.client.api,library=jvm-retrofit2,useCoroutines=true,dateLibrary=kotlinx-datetime

echo "✓ API clients generated successfully."
