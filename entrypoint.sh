#!/bin/sh
set -e

# Build command arguments
ARGS="check"

# Add directory argument
ARGS="$ARGS --dir ${INPUT_DIRECTORY:-./}"

# Add verbose flag if enabled
if [ "$INPUT_VERBOSE" = "true" ]; then
    ARGS="$ARGS --verbose"
fi

# Add output JSON flag if enabled
if [ "$INPUT_OUTPUT_JSON" = "true" ]; then
    ARGS="$ARGS --output-json"
fi

# Add output file if specified
if [ -n "$INPUT_OUTPUT_FILE" ]; then
    ARGS="$ARGS --output-file $INPUT_OUTPUT_FILE"
fi

# Add teams webhook flag if webhook is provided
if [ -n "$SUPPRESS_TEAMS_WEBHOOK" ]; then
    ARGS="$ARGS --teams"
fi

# Add dry-run flag if enabled
if [ "$INPUT_DRY_RUN" = "true" ]; then
    ARGS="$ARGS --dry-run"
fi

# Add fail-on-warnings flag if enabled
if [ "$INPUT_FAIL_ON_WARNINGS" = "true" ]; then
    ARGS="$ARGS --fail-on-warnings"
fi

# Add grace period if different from default
if [ -n "$INPUT_GRACE_PERIOD" ] && [ "$INPUT_GRACE_PERIOD" != "30" ]; then
    ARGS="$ARGS --grace-period $INPUT_GRACE_PERIOD"
fi

echo "Running: suppress-checker $ARGS"

# Execute the command
exec suppress-checker $ARGS 