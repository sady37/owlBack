#!/bin/bash
#
# wisefido-sleepace API Test Script
# Test various sleepace API endpoints
#

set -e

# Default values
SERVER="http://127.0.0.1:8083"
VERBOSE=0
INTERFACE=""
DEVICE_CODE=""
DEVICE_ID=""
USER_CODE=""
TENANT_ID=""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Test wisefido-sleepace API endpoints

OPTIONS:
    -s <server>       Sleepace server address (default: http://127.0.0.1:8083)
    -v                Verbose mode - show request and response JSON
    -i <interface>    Specify API interface to test (see below)
    -c <device_code>  Sleepace device code (deviceId) for testing
    -d <device_id>    Platform device UUID for initialize
    -u <user_code>    User code for device binding
    -t <tenant_id>    Tenant ID for batch queries
    -h                Show this help message

AVAILABLE INTERFACES:
    health             Health check endpoint
    status             Get single device online status
    status-batch       Get batch device status (optional: -t tenant_id)
    initialize         Initialize/bind device (requires: -c device_code -d device_id -u user_code)
    unbind             Unbind device (requires: -c device_code)
    alarm-get          Get alarm notification config (requires: -c device_code)
    alarm-set          Update alarm config (requires: -c device_code -a alarm_json)
    realtime-interval  Set realtime data interval (requires: -c device_code -r interval)
    heart-mode         Set heart rate mode (requires: -c device_code -m mode)
    leave-sensitivity  Set leave bed sensitivity (requires: -c device_code -l alg_mode)
    mattress-params    Set mattress parameters (requires: -c device_code -p params_json)
    sleep-report       Get sleep report (requires: -c device_code -e report_date)

EXAMPLES:
    # Health check
    $0 -i health

    # Get device status
    $0 -i status -c 1ua3erivl9pv1

    # Initialize device
    $0 -i initialize -c 1ua3erivl9pv1 -d 550e8400-e29b-41d4-a716-446655440000 -u user123

    # Unbind device
    $0 -i unbind -c 1ua3erivl9pv1

    # Get alarm config
    $0 -i alarm-get -c 1ua3erivl9pv1 -v

    # Update alarm config
    $0 -i alarm-set -c 1ua3erivl9pv1 -a '[{"alarmType":"HEARTRATE_HIGH","isEnabled":1,"threshold":120}]'

    # Set realtime interval to 30 seconds
    $0 -i realtime-interval -c 1ua3erivl9pv1 -r 30

    # Enable heart rate monitoring
    $0 -i heart-mode -c 1ua3erivl9pv1 -m 1

    # Set leave sensitivity (0=low, 1=medium, 2=high)
    $0 -i leave-sensitivity -c 1ua3erivl9pv1 -l 1

    # Set mattress parameters
    $0 -i mattress-params -c 1ua3erivl9pv1 -p '{"mattressThickness":20,"mattressMaterial":"memory_foam"}'

    # Get sleep report for 2026-04-02
    $0 -i sleep-report -c 1ua3erivl9pv1 -e 20260402

    # Batch device status
    $0 -i status-batch

EOF
    exit 0
}

# Print colored message
print_msg() {
    local color=$1
    shift
    echo -e "${color}$@${NC}"
}

# Print request details
print_request() {
    local method=$1
    local url=$2
    local body=$3

    if [ "$VERBOSE" -eq 1 ]; then
        print_msg "$BLUE" "\n==> REQUEST"
        echo "Method: $method"
        echo "URL: $url"
        if [ -n "$body" ]; then
            echo "Body:"
            echo "$body" | jq '.' 2>/dev/null || echo "$body"
        fi
    fi
}

# Print response details
print_response() {
    local status=$1
    local response=$2

    if [ "$VERBOSE" -eq 1 ]; then
        print_msg "$BLUE" "\n==> RESPONSE"
        echo "Status: $status"
        echo "Body:"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    fi

    # Print summary
    local success=$(echo "$response" | jq -r '.success' 2>/dev/null || echo "unknown")
    if [ "$success" = "true" ]; then
        print_msg "$GREEN" "✓ SUCCESS"
    elif [ "$success" = "false" ]; then
        print_msg "$RED" "✗ FAILED"
        local error=$(echo "$response" | jq -r '.error' 2>/dev/null || echo "")
        if [ -n "$error" ] && [ "$error" != "null" ]; then
            print_msg "$RED" "Error: $error"
        fi
    else
        print_msg "$YELLOW" "? UNKNOWN RESPONSE FORMAT"
    fi
}

# Test health endpoint
test_health() {
    local url="${SERVER}/health"
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test get device status
test_status() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/sleepace/devices/${DEVICE_CODE}/status"
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test batch device status
test_status_batch() {
    local url="${SERVER}/api/v1/sleepace/devices/status"
    if [ -n "$TENANT_ID" ]; then
        url="${url}?tenant_id=${TENANT_ID}"
    fi
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test initialize device
test_initialize() {
    if [ -z "$DEVICE_CODE" ] || [ -z "$DEVICE_ID" ] || [ -z "$USER_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code>, -d <device_id>, and -u <user_code> are required"
        exit 1
    fi

    local url="${SERVER}/api/v1/sleepace/device/initialize"
    local body="{\"device_code\":\"${DEVICE_CODE}\",\"device_id\":\"${DEVICE_ID}\",\"user_code\":\"${USER_CODE}\"}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test unbind device
test_unbind() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/sleepace/device/${DEVICE_CODE}"
    print_request "DELETE" "$url"

    local response=$(curl -s -w "\n%{http_code}" -X DELETE "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test get alarm config
test_alarm_get() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/getalarmnotifyconfig"
    local body="{\"deviceId\":\"${DEVICE_CODE}\"}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test update alarm config
test_alarm_set() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$ALARM_JSON" ]; then
        print_msg "$RED" "Error: -a <alarm_json> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/updatealarmnotifyconfig"
    local body="{\"deviceId\":\"${DEVICE_CODE}\",\"alarmNotifySettings\":${ALARM_JSON}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test set realtime interval
test_realtime_interval() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$REALTIME_INTERVAL" ]; then
        print_msg "$RED" "Error: -r <interval> is required (10-60 seconds)"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/device/updateconfig"
    local body="{\"deviceId\":\"${DEVICE_CODE}\",\"realtimeInterval\":${REALTIME_INTERVAL}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test set heart mode
test_heart_mode() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$HEART_MODE" ]; then
        print_msg "$RED" "Error: -m <mode> is required (0=off, 1=on)"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/heartModeSet"
    local body="{\"deviceId\":\"${DEVICE_CODE}\",\"heartMode\":${HEART_MODE}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test set leave sensitivity
test_leave_sensitivity() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$ALG_MODE" ]; then
        print_msg "$RED" "Error: -l <alg_mode> is required (0=low, 1=medium, 2=high)"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/device/updateAlgMode"
    local body="{\"deviceId\":\"${DEVICE_CODE}\",\"algMode\":${ALG_MODE}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test set mattress parameters
test_mattress_params() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$PARAMS_JSON" ]; then
        print_msg "$RED" "Error: -p <params_json> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/updateSetting"
    local params=$(echo "$PARAMS_JSON" | jq -c .)
    local body="{\"deviceId\":\"${DEVICE_CODE}\",$(echo "$params" | sed 's/^{//;s/}$//')}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test get sleep report
test_sleep_report() {
    if [ -z "$DEVICE_CODE" ]; then
        print_msg "$RED" "Error: -c <device_code> is required"
        exit 1
    fi
    if [ -z "$REPORT_DATE" ]; then
        print_msg "$RED" "Error: -e <report_date> is required (format: YYYYMMDD)"
        exit 1
    fi

    local url="${SERVER}/api/v1/proxy/sleepace/get24HourDailyWithMaxReport"
    local body="{\"deviceId\":\"${DEVICE_CODE}\",\"reportDate\":\"${REPORT_DATE}\"}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Parse command line arguments
while getopts "s:vi:c:d:u:t:a:r:m:l:p:e:h" opt; do
    case $opt in
        s)
            SERVER="$OPTARG"
            ;;
        v)
            VERBOSE=1
            ;;
        i)
            INTERFACE="$OPTARG"
            ;;
        c)
            DEVICE_CODE="$OPTARG"
            ;;
        d)
            DEVICE_ID="$OPTARG"
            ;;
        u)
            USER_CODE="$OPTARG"
            ;;
        t)
            TENANT_ID="$OPTARG"
            ;;
        a)
            ALARM_JSON="$OPTARG"
            ;;
        r)
            REALTIME_INTERVAL="$OPTARG"
            ;;
        m)
            HEART_MODE="$OPTARG"
            ;;
        l)
            ALG_MODE="$OPTARG"
            ;;
        p)
            PARAMS_JSON="$OPTARG"
            ;;
        e)
            REPORT_DATE="$OPTARG"
            ;;
        h)
            usage
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            usage
            ;;
    esac
done

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    print_msg "$YELLOW" "Warning: jq is not installed. Output formatting will be limited."
fi

# Main execution
if [ -z "$INTERFACE" ]; then
    print_msg "$RED" "Error: -i <interface> is required"
    echo ""
    usage
fi

print_msg "$BLUE" "Testing wisefido-sleepace API"
print_msg "$BLUE" "Server: $SERVER"
print_msg "$BLUE" "Interface: $INTERFACE"
if [ -n "$DEVICE_CODE" ]; then
    print_msg "$BLUE" "Device Code: $DEVICE_CODE"
fi
echo ""

case "$INTERFACE" in
    health)
        test_health
        ;;
    status)
        test_status
        ;;
    status-batch)
        test_status_batch
        ;;
    initialize)
        test_initialize
        ;;
    unbind)
        test_unbind
        ;;
    alarm-get)
        test_alarm_get
        ;;
    alarm-set)
        test_alarm_set
        ;;
    realtime-interval)
        test_realtime_interval
        ;;
    heart-mode)
        test_heart_mode
        ;;
    leave-sensitivity)
        test_leave_sensitivity
        ;;
    mattress-params)
        test_mattress_params
        ;;
    sleep-report)
        test_sleep_report
        ;;
    *)
        print_msg "$RED" "Error: Unknown interface '$INTERFACE'"
        echo ""
        usage
        ;;
esac

echo ""
