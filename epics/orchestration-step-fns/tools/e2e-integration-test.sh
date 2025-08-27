#!/bin/bash
# e2e-integration-test.sh - End-to-end integration test for Step Functions workflow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default configuration
DEFAULT_REGION="us-east-1"
DEFAULT_PROFILE="default"
DEFAULT_ENVIRONMENT="mock"

# Configuration from environment or defaults
REGION="${AWS_REGION:-$DEFAULT_REGION}"
PROFILE="${AWS_PROFILE:-$DEFAULT_PROFILE}"
ENVIRONMENT="${ENVIRONMENT:-$DEFAULT_ENVIRONMENT}"
PROJECT_PREFIX="${PROJECT_PREFIX:-jaunt}"

# Test configuration
TEST_TIMEOUT=1800  # 30 minutes
CHECK_INTERVAL=30  # 30 seconds
VERBOSE=${VERBOSE:-false}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
End-to-End Integration Test for Step Functions Workflow

Usage: $0 [OPTIONS]

Options:
  --region REGION         AWS region (default: $DEFAULT_REGION)
  --profile PROFILE       AWS profile (default: $DEFAULT_PROFILE)
  --env ENVIRONMENT       Environment (default: $DEFAULT_ENVIRONMENT)
  --project-prefix PREFIX Project prefix (default: $PROJECT_PREFIX)
  --timeout SECONDS       Test timeout in seconds (default: 1800)
  --interval SECONDS      Check interval in seconds (default: 30)
  --verbose               Enable verbose output
  --city CITY             Test city (default: edinburgh)
  --fail-fast BOOL        Override orchestrator.fail_fast (true|false)
  --help                  Show this help message

Examples:
  $0                                    # Run with defaults
  $0 --env dev --city london           # Test dev environment with London
  $0 --verbose --timeout 3600          # Extended timeout with verbose output

EOF
}

log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case $level in
        INFO)
            echo -e "${BLUE}[$timestamp] INFO: $message${NC}" >&2
            ;;
        WARN)
            echo -e "${YELLOW}[$timestamp] WARN: $message${NC}" >&2
            ;;
        ERROR)
            echo -e "${RED}[$timestamp] ERROR: $message${NC}" >&2
            ;;
        SUCCESS)
            echo -e "${GREEN}[$timestamp] SUCCESS: $message${NC}" >&2
            ;;
        DEBUG)
            if [ "$VERBOSE" = true ]; then
                echo -e "[$timestamp] DEBUG: $message" >&2
            fi
            ;;
    esac
}

check_prerequisites() {
    log INFO "Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        log ERROR "AWS CLI not found. Please install aws cli."
        exit 1
    fi
    
    # Check jq
    if ! command -v jq &> /dev/null; then
        log ERROR "jq not found. Please install jq for JSON processing."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity --profile "$PROFILE" --region "$REGION" &> /dev/null; then
        log ERROR "AWS credentials not configured properly for profile: $PROFILE"
        exit 1
    fi
    
    # Get AWS account ID
    AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text --profile "$PROFILE" --region "$REGION")
    log INFO "Using AWS Account: $AWS_ACCOUNT"
    
    log SUCCESS "Prerequisites check completed"
}

check_infrastructure() {
    log INFO "Checking infrastructure deployment..."
    
    # Check State Machine
    STATE_MACHINE_ARN="arn:aws:states:${REGION}:${AWS_ACCOUNT}:stateMachine:data-scout-orchestration-step-function"
    if ! aws stepfunctions describe-state-machine \
        --state-machine-arn "$STATE_MACHINE_ARN" \
        --profile "$PROFILE" \
        --region "$REGION" &> /dev/null; then
        log ERROR "State machine not found: $STATE_MACHINE_ARN"
        log ERROR "Please deploy infrastructure first: cd terraform && make apply ENV=$ENVIRONMENT"
        exit 1
    fi
    
    # Resolve SQS queue URLs dynamically by name
    FRONTIER_NAME="${PROJECT_PREFIX}-${ENVIRONMENT}-frontier"
    DLQ_NAME="${PROJECT_PREFIX}-${ENVIRONMENT}-frontier-dlq"
    log INFO "Expected frontier queue name: $FRONTIER_NAME"
    log INFO "Expected DLQ name: $DLQ_NAME"

    FRONTIER_URL=$(aws sqs get-queue-url \
        --queue-name "$FRONTIER_NAME" \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'QueueUrl' \
        --output text 2>/dev/null || true)

    if [ -z "$FRONTIER_URL" ] || [ "$FRONTIER_URL" = "None" ] || [ "$FRONTIER_URL" = "null" ]; then
        log ERROR "Queue not found: $FRONTIER_NAME in region $REGION (account: $AWS_ACCOUNT)"
        log ERROR "Fix: deploy with matching flags (e.g., cd terraform && make apply ENV=$ENVIRONMENT PROJECT_PREFIX=$PROJECT_PREFIX)"
        exit 1
    fi

    DLQ_URL=$(aws sqs get-queue-url \
        --queue-name "$DLQ_NAME" \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'QueueUrl' \
        --output text 2>/dev/null || true)

    if [ -z "$DLQ_URL" ] || [ "$DLQ_URL" = "None" ] || [ "$DLQ_URL" = "null" ]; then
        log ERROR "Queue not found: $DLQ_NAME in region $REGION (account: $AWS_ACCOUNT)"
        log ERROR "Fix: deploy with matching flags (e.g., cd terraform && make apply ENV=$ENVIRONMENT PROJECT_PREFIX=$PROJECT_PREFIX)"
        exit 1
    fi

    log INFO "Resolved frontier queue URL: $FRONTIER_URL"
    log INFO "Resolved DLQ URL: $DLQ_URL"
    
    # Check S3 bucket
    S3_BUCKET="${PROJECT_PREFIX}-${ENVIRONMENT}-data-scout-raw-${AWS_ACCOUNT}"
    if ! aws s3api head-bucket --bucket "$S3_BUCKET" --profile "$PROFILE" --region "$REGION" &> /dev/null; then
        log WARN "S3 bucket not found: $S3_BUCKET (this may be expected for some configurations)"
    fi
    
    log SUCCESS "Infrastructure check completed"
}

prepare_test_input() {
    local city=${1:-edinburgh}
    log INFO "Preparing test input for city: $city"
    
    INPUT_FILE="$SCRIPT_DIR/../examples/input.${city}.json"
    if [ ! -f "$INPUT_FILE" ]; then
        log ERROR "Input file not found: $INPUT_FILE"
        log ERROR "Available inputs: $(ls $SCRIPT_DIR/../examples/input.*.json | xargs basename -s .json | sed 's/input\.//' | tr '\n' ' ')"
        exit 1
    fi
    
    # Validate JSON
    if ! jq . "$INPUT_FILE" > /dev/null; then
        log ERROR "Invalid JSON in input file: $INPUT_FILE"
        exit 1
    fi

    # Normalize payload: ensure orchestrator exists and has fail_fast (default false or CLI override)
    PREPARED_PAYLOAD=$(jq -c --arg fail_fast "${FAIL_FAST_OPT:-}" '
      .orchestrator = (.orchestrator // {}) |
      .orchestrator.fail_fast =
        (if $fail_fast == "" then
            (.orchestrator.fail_fast // false)
         else
            # parse string to boolean
            (if ($fail_fast|test("^(?i:true)$")) then true else false end)
         end)
    ' "$INPUT_FILE") || {
        log ERROR "Failed to prepare payload from: $INPUT_FILE"
        exit 1
    }

    if [ -z "$PREPARED_PAYLOAD" ] || [ "$PREPARED_PAYLOAD" = "null" ]; then
        log ERROR "Prepared payload is empty"
        exit 1
    fi
    
    log DEBUG "Using input file: $INPUT_FILE"
    log SUCCESS "Test input prepared"
}

start_execution() {
    local city=${1:-edinburgh}
    local execution_name="e2e-test-${city}-$(date +%s)"
    
    log INFO "Starting Step Functions execution: $execution_name"
    
    EXECUTION_ARN=$(aws stepfunctions start-execution \
        --state-machine-arn "$STATE_MACHINE_ARN" \
        --name "$execution_name" \
        --input "$PREPARED_PAYLOAD" \
        --cli-binary-format raw-in-base64-out \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'executionArn' \
        --output text)
    
    if [ $? -ne 0 ]; then
        log ERROR "Failed to start execution"
        exit 1
    fi
    
    log SUCCESS "Execution started: $EXECUTION_ARN"
    echo "$EXECUTION_ARN"
}

monitor_execution() {
    local execution_arn=$1
    local start_time=$(date +%s)
    local end_time=$((start_time + TEST_TIMEOUT))
    
    log INFO "Monitoring execution (timeout: ${TEST_TIMEOUT}s, check interval: ${CHECK_INTERVAL}s)"
    log INFO "Execution ARN: $execution_arn"
    
    while [ $(date +%s) -lt $end_time ]; do
        local status=$(aws stepfunctions describe-execution \
            --execution-arn "$execution_arn" \
            --profile "$PROFILE" \
            --region "$REGION" \
            --query 'status' \
            --output text)
        
        case $status in
            SUCCEEDED)
                log SUCCESS "Execution completed successfully"
                return 0
                ;;
            FAILED)
                log ERROR "Execution failed"
                show_execution_details "$execution_arn"
                return 1
                ;;
            TIMED_OUT)
                log ERROR "Execution timed out"
                show_execution_details "$execution_arn"
                return 1
                ;;
            ABORTED)
                log ERROR "Execution was aborted"
                show_execution_details "$execution_arn"
                return 1
                ;;
            RUNNING)
                local elapsed=$(($(date +%s) - start_time))
                log INFO "Execution still running (elapsed: ${elapsed}s)..."
                
                if [ "$VERBOSE" = true ]; then
                    show_current_state "$execution_arn"
                fi
                ;;
            *)
                log WARN "Unknown execution status: $status"
                ;;
        esac
        
        sleep $CHECK_INTERVAL
    done
    
    log ERROR "Execution monitoring timed out after ${TEST_TIMEOUT} seconds"
    return 1
}

show_current_state() {
    local execution_arn=$1
    
    local current_state=$(aws stepfunctions get-execution-history \
        --execution-arn "$execution_arn" \
        --profile "$PROFILE" \
        --region "$REGION" \
        --max-results 1 \
        --reverse-order \
        --query 'events[0].stateEnteredEventDetails.name' \
        --output text)
    
    if [ "$current_state" != "null" ] && [ -n "$current_state" ]; then
        log DEBUG "Current state: $current_state"
    fi
}

show_execution_details() {
    local execution_arn=$1
    
    log INFO "Execution details:"
    aws stepfunctions describe-execution \
        --execution-arn "$execution_arn" \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query '{status: status, startDate: startDate, stopDate: stopDate, error: error, cause: cause}'
    
    if [ "$VERBOSE" = true ]; then
        log INFO "Execution history (last 10 events):"
        aws stepfunctions get-execution-history \
            --execution-arn "$execution_arn" \
            --profile "$PROFILE" \
            --region "$REGION" \
            --max-results 10 \
            --reverse-order \
            --query 'events[].{timestamp: timestamp, type: type, stateEnteredEventDetails: stateEnteredEventDetails, taskFailedEventDetails: taskFailedEventDetails}'

        # Try to surface the input to the Choice state 'ToDLQOrContinue' if logging captured it
        local choice_input
        choice_input=$(aws stepfunctions get-execution-history \
            --execution-arn "$execution_arn" \
            --profile "$PROFILE" \
            --region "$REGION" \
            --max-results 50 \
            --reverse-order \
            --output json 2>/dev/null | jq -r '
              (.events[] | select(.type=="ChoiceStateEntered" and .stateEnteredEventDetails.name=="ToDLQOrContinue") | .stateEnteredEventDetails.input) // empty
            ' || true)
        if [ -n "$choice_input" ]; then
            log INFO "Input at 'ToDLQOrContinue' (from history):"
            echo "$choice_input" | jq . >&2 || echo "$choice_input" >&2
        else
            log DEBUG "State input not available; enable execution data logging on the state machine to capture it."
        fi
    fi
}

check_outputs() {
    local execution_arn=$1
    
    log INFO "Checking execution outputs..."
    
    # Get execution output
    local output=$(aws stepfunctions describe-execution \
        --execution-arn "$execution_arn" \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'output' \
        --output text)
    
    if [ "$output" != "null" ] && [ -n "$output" ]; then
        log DEBUG "Execution output: $output"
        
        # Validate output structure
        if echo "$output" | jq . > /dev/null 2>&1; then
            log SUCCESS "Valid JSON output received"
        else
            log WARN "Output is not valid JSON"
        fi
    else
        log WARN "No execution output received"
    fi
    
    # Check SQS queue states
    check_queue_states
    
    # Check S3 outputs (if bucket exists)
    check_s3_outputs
}

check_queue_states() {
    log INFO "Checking SQS queue states..."
    
    # Check frontier queue
    local frontier_msgs=$(aws sqs get-queue-attributes \
        --queue-url "$FRONTIER_URL" \
        --attribute-names ApproximateNumberOfMessages \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'Attributes.ApproximateNumberOfMessages' \
        --output text)
    
    # Check DLQ
    local dlq_msgs=$(aws sqs get-queue-attributes \
        --queue-url "$DLQ_URL" \
        --attribute-names ApproximateNumberOfMessages \
        --profile "$PROFILE" \
        --region "$REGION" \
        --query 'Attributes.ApproximateNumberOfMessages' \
        --output text)
    
    log INFO "Frontier queue messages: $frontier_msgs"
    log INFO "DLQ messages: $dlq_msgs"
    
    if [ "$dlq_msgs" -gt 0 ]; then
        log WARN "$dlq_msgs messages found in DLQ - check for processing failures"
        if [ "$VERBOSE" = true ]; then
            log INFO "DLQ message sample:"
            aws sqs receive-message \
                --queue-url "$DLQ_URL" \
                --max-number-of-messages 1 \
                --profile "$PROFILE" \
                --region "$REGION" \
                --query 'Messages[0].Body'
        fi
    fi
}

check_s3_outputs() {
    log INFO "Checking S3 outputs..."
    
    if aws s3api head-bucket --bucket "$S3_BUCKET" --profile "$PROFILE" --region "$REGION" &> /dev/null; then
        local object_count=$(aws s3 ls "s3://$S3_BUCKET/" --recursive --profile "$PROFILE" --region "$REGION" | wc -l)
        log INFO "S3 bucket objects: $object_count"
        
        if [ "$VERBOSE" = true ] && [ "$object_count" -gt 0 ]; then
            log INFO "Recent S3 objects:"
            aws s3 ls "s3://$S3_BUCKET/" --recursive --profile "$PROFILE" --region "$REGION" | tail -5
        fi
    else
        log DEBUG "S3 bucket not accessible or doesn't exist: $S3_BUCKET"
    fi
}

generate_test_report() {
    local execution_arn=$1
    local test_result=$2
    local city=$3
    
    log INFO "Generating test report..."
    
    local report_file="e2e-test-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "test_info": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "environment": "$ENVIRONMENT",
    "region": "$REGION",
    "city": "$city",
    "result": "$test_result"
  },
  "execution": {
    "arn": "$execution_arn",
    "details": $(aws stepfunctions describe-execution --execution-arn "$execution_arn" --profile "$PROFILE" --region "$REGION" 2>/dev/null || echo 'null')
  },
  "infrastructure": {
    "state_machine_arn": "$STATE_MACHINE_ARN",
    "frontier_queue_url": "$FRONTIER_URL",
    "dlq_url": "$DLQ_URL",
    "s3_bucket": "$S3_BUCKET"
  }
}
EOF
    
    log SUCCESS "Test report generated: $report_file"
}

cleanup() {
    local execution_arn=${1:-}
    
    if [ -n "$execution_arn" ]; then
        log INFO "Checking if execution cleanup is needed..."
        
        local status=$(aws stepfunctions describe-execution \
            --execution-arn "$execution_arn" \
            --profile "$PROFILE" \
            --region "$REGION" \
            --query 'status' \
            --output text 2>/dev/null || echo "UNKNOWN")
        
        if [ "$status" = "RUNNING" ]; then
            log WARN "Execution still running. Consider stopping it manually if needed."
            log INFO "Stop command: aws stepfunctions stop-execution --execution-arn $execution_arn"
        fi
    fi
    
    log INFO "Cleanup completed"
}

run_integration_test() {
    local city=${1:-edinburgh}
    
    log INFO "Starting end-to-end integration test"
    log INFO "Environment: $ENVIRONMENT, City: $city, Region: $REGION"
    
    # Check prerequisites
    check_prerequisites
    
    # Check infrastructure
    check_infrastructure
    
    # Prepare test input
    prepare_test_input "$city"
    
    # Start execution
    local execution_arn
    execution_arn=$(start_execution "$city")
    
    # Monitor execution
    local test_result
    if monitor_execution "$execution_arn"; then
        test_result="PASSED"
        log SUCCESS "Integration test PASSED"
        
        # Check outputs
        check_outputs "$execution_arn"
    else
        test_result="FAILED"
        log ERROR "Integration test FAILED"
    fi
    
    # Generate report
    generate_test_report "$execution_arn" "$test_result" "$city"
    
    # Cleanup
    cleanup "$execution_arn"
    
    return $([ "$test_result" = "PASSED" ] && echo 0 || echo 1)
}

# Parse command line arguments
CITY="edinburgh"
FAIL_FAST_OPT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --region)
            REGION="$2"
            shift 2
            ;;
        --profile)
            PROFILE="$2"
            shift 2
            ;;
        --env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --project-prefix)
            PROJECT_PREFIX="$2"
            shift 2
            ;;
        --timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --interval)
            CHECK_INTERVAL="$2"
            shift 2
            ;;
        --city)
            CITY="$2"
            shift 2
            ;;
        --fail-fast)
            FAIL_FAST_OPT="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Trap for cleanup on exit
trap 'cleanup' EXIT

# Run the integration test
run_integration_test "$CITY"