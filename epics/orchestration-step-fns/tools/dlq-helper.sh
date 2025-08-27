#!/bin/bash
# dlq-helper.sh - DLQ management utility

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_REGION="us-east-1"
DEFAULT_PROFILE="default"

# Configuration
REGION="${AWS_REGION:-$DEFAULT_REGION}"
PROFILE="${AWS_PROFILE:-$DEFAULT_PROFILE}"
PROJECT_PREFIX="${PROJECT_PREFIX:-jaunt}"
ENVIRONMENT="${ENVIRONMENT:-dev}"

usage() {
    cat << EOF
DLQ Helper Tool

Usage: $0 [OPTIONS] COMMAND

Commands:
  status          Show DLQ status and message count
  peek            Peek at messages without consuming them
  analyze         Analyze message patterns and common errors
  redrive-all     Move all messages from DLQ back to frontier queue
  redrive-batch   Move specified number of messages back to frontier queue
  purge           Delete all messages from DLQ (use with caution)
  monitor         Monitor DLQ in real-time

Options:
  --region REGION     AWS region (default: $DEFAULT_REGION)
  --profile PROFILE   AWS profile (default: $DEFAULT_PROFILE)
  --env ENVIRONMENT   Environment (default: dev)
  --help              Show this help message

Examples:
  $0 status
  $0 peek --count 5
  $0 redrive-batch --count 10
  $0 monitor --interval 30

EOF
}

status() {
    echo "DLQ Status for $ENVIRONMENT environment"
    echo "======================================="
    
    ATTRS=$(aws sqs get-queue-attributes \
        --queue-url "$DLQ_URL" \
        --attribute-names ApproximateNumberOfMessages,ApproximateNumberOfMessagesNotVisible,ApproximateNumberOfMessagesDelayed \
        --region "$REGION" \
        --profile "$PROFILE")
    
    VISIBLE=$(echo "$ATTRS" | jq -r '.Attributes.ApproximateNumberOfMessages')
    NOT_VISIBLE=$(echo "$ATTRS" | jq -r '.Attributes.ApproximateNumberOfMessagesNotVisible')
    DELAYED=$(echo "$ATTRS" | jq -r '.Attributes.ApproximateNumberOfMessagesDelayed')
    
    echo "Messages available: $VISIBLE"
    echo "Messages in flight: $NOT_VISIBLE"
    echo "Messages delayed: $DELAYED"
    echo "Total messages: $((VISIBLE + NOT_VISIBLE + DELAYED))"
}

peek() {
    local count=${1:-5}
    echo "Peeking at $count messages from DLQ..."
    
    aws sqs receive-message \
        --queue-url "$DLQ_URL" \
        --max-number-of-messages "$count" \
        --region "$REGION" \
        --profile "$PROFILE" | \
    jq -r '.Messages[]? | "Message ID: " + .MessageId + "\nBody: " + .Body + "\n---"'
}

analyze() {
    echo "Analyzing DLQ messages..."
    
    # Get sample of messages
    MESSAGES=$(aws sqs receive-message \
        --queue-url "$DLQ_URL" \
        --max-number-of-messages 10 \
        --attribute-names All \
        --region "$REGION" \
        --profile "$PROFILE")
    
    # Analyze patterns
    echo "Message types:"
    echo "$MESSAGES" | jq -r '.Messages[]?.Body' | jq -r '.type // "unknown"' | sort | uniq -c
    
    echo -e "\nCorrelation IDs:"
    echo "$MESSAGES" | jq -r '.Messages[]?.Body' | jq -r '.correlation_id // "none"' | sort | uniq -c
    
    echo -e "\nJob IDs:"
    echo "$MESSAGES" | jq -r '.Messages[]?.Body' | jq -r '.job_id // "none"' | sort | uniq -c
}

redrive_all() {
    echo "Moving all messages from DLQ back to frontier queue..."
    
    aws sqs start-message-move-task \
        --source-arn "arn:aws:sqs:${REGION}:${AWS_ACCOUNT}:${PROJECT_PREFIX}-${ENVIRONMENT}-frontier-dlq" \
        --destination-arn "arn:aws:sqs:${REGION}:${AWS_ACCOUNT}:${PROJECT_PREFIX}-${ENVIRONMENT}-frontier" \
        --region "$REGION" \
        --profile "$PROFILE"
    
    echo "Redrive task started. Monitor with: aws sqs list-message-move-tasks"
}

redrive_batch() {
    local count=${1:-10}
    echo "Moving $count messages from DLQ back to frontier queue..."
    
    for i in $(seq 1 "$count"); do
        # Get one message
        MESSAGE=$(aws sqs receive-message \
            --queue-url "$DLQ_URL" \
            --max-number-of-messages 1 \
            --region "$REGION" \
            --profile "$PROFILE")
        
        if [ "$(echo "$MESSAGE" | jq '.Messages | length')" -eq 0 ]; then
            echo "No more messages in DLQ"
            break
        fi
        
        BODY=$(echo "$MESSAGE" | jq -r '.Messages[0].Body')
        RECEIPT_HANDLE=$(echo "$MESSAGE" | jq -r '.Messages[0].ReceiptHandle')
        
        # Send to frontier queue
        aws sqs send-message \
            --queue-url "$FRONTIER_URL" \
            --message-body "$BODY" \
            --region "$REGION" \
            --profile "$PROFILE" > /dev/null
        
        # Delete from DLQ
        aws sqs delete-message \
            --queue-url "$DLQ_URL" \
            --receipt-handle "$RECEIPT_HANDLE" \
            --region "$REGION" \
            --profile "$PROFILE"
        
        echo "Processed message $i/$count"
        sleep 0.1  # Rate limiting
    done
    
    echo "Batch redrive completed"
}

monitor() {
    local interval=${1:-30}
    echo "Monitoring DLQ (checking every ${interval} seconds, Ctrl+C to stop)..."
    
    while true; do
        clear
        echo "$(date): DLQ Monitor"
        status
        echo ""
        echo "Press Ctrl+C to stop monitoring"
        sleep "$interval"
    done
}

# Parse command line arguments
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
        --help)
            usage
            exit 0
            ;;
        status|peek|analyze|redrive-all|redrive-batch|purge|monitor)
            COMMAND="$1"
            shift
            ;;
        --count)
            COUNT="$2"
            shift 2
            ;;
        --interval)
            INTERVAL="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Get AWS account ID
AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text --region "$REGION" --profile "$PROFILE")

# Update queue URLs with actual account ID
DLQ_URL="https://sqs.${REGION}.amazonaws.com/${AWS_ACCOUNT}/${PROJECT_PREFIX}-${ENVIRONMENT}-frontier-dlq"
FRONTIER_URL="https://sqs.${REGION}.amazonaws.com/${AWS_ACCOUNT}/${PROJECT_PREFIX}-${ENVIRONMENT}-frontier"

# Execute command
case $COMMAND in
    status)
        status
        ;;
    peek)
        peek "${COUNT:-5}"
        ;;
    analyze)
        analyze
        ;;
    redrive-all)
        redrive_all
        ;;
    redrive-batch)
        redrive_batch "${COUNT:-10}"
        ;;
    monitor)
        monitor "${INTERVAL:-30}"
        ;;
    *)
        echo "No command specified"
        usage
        exit 1
        ;;
esac