#!/bin/bash
#
# wisefido-qinglan API Test Script
# Test various qinglan API endpoints
#

set -e

# Default values
SERVER="http://127.0.0.1:8081"
VERBOSE=0
INTERFACE=""
DEVICE_UID=""

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

Test wisefido-qinglan API endpoints

OPTIONS:
    -s <server>     Qinglan server address (default: http://127.0.0.1:8081)
    -v              Verbose mode - show request and response JSON
    -i <interface>  Specify API interface to test (see below)
    -u <device_uid> Device UID for testing (required for most APIs)
    -h              Show this help message

AVAILABLE INTERFACES:
    health          Health check endpoint
    status          Get device online status
    status-batch    Get batch device status (optional: -t tenant_id)
    info            Get device metadata from database
    properties      Read device properties (optional: -k keys)
    properties-set  Write device properties (requires: -p properties_json)
    subscribe       Start realtime data subscription (optional: -c content -d duration)
    function        Call device function (requires: -f dev_code)
    tenant-devices  Get devices by tenant (requires: -t tenant_id)

EXAMPLES:
    # Health check
    $0 -i health

    # Get device status
    $0 -i status -u BM87XXXX

    # Read all device properties
    $0 -i properties -u BM87XXXX -v

    # Read specific properties
    $0 -i properties -u BM87XXXX -k "install_height,heart_breath_switch"

    # Write device properties
    $0 -i properties-set -u BM87XXXX -p '{"install_height":280,"install_model":0}'

    # Restart device (dev: 0=all, 1=radar, 2=MCU)
    $0 -i function -u BM87XXXX -f 0

    # Subscribe realtime data (content: 0=all, 1=track, 2=vital; duration: seconds)
    $0 -i subscribe -u BM87XXXX -c 0 -d 300

    # Batch device status
    $0 -i status-batch

    # Get tenant devices
    $0 -i tenant-devices -t tenant-uuid

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
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/status"
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test batch device status
test_status_batch() {
    local url="${SERVER}/api/v1/radar/devices/status"
    if [ -n "$TENANT_ID" ]; then
        url="${url}?tenant_id=${TENANT_ID}"
    fi
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test get device info
test_info() {
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/info"
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test read device properties
test_properties() {
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/properties"
    if [ -n "$KEYS" ]; then
        url="${url}?keys=${KEYS}"
    fi
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test write device properties
test_properties_set() {
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi
    if [ -z "$PROPERTIES_JSON" ]; then
        print_msg "$RED" "Error: -p <properties_json> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/properties"
    local body="{\"properties\":${PROPERTIES_JSON}}"
    print_request "PUT" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X PUT "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test subscribe realtime data
test_subscribe() {
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi

    local content=${CONTENT:-0}
    local duration=${DURATION:-300}

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/subscribe"
    local body="{\"content\":${content},\"duration\":${duration}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test call device function
test_function() {
    if [ -z "$DEVICE_UID" ]; then
        print_msg "$RED" "Error: -u <device_uid> is required"
        exit 1
    fi
    if [ -z "$DEV_CODE" ]; then
        print_msg "$RED" "Error: -f <dev_code> is required (0=all, 1=radar, 2=MCU)"
        exit 1
    fi

    local url="${SERVER}/api/v1/radar/devices/${DEVICE_UID}/function"
    local body="{\"dev\":${DEV_CODE}}"
    print_request "POST" "$url" "$body"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Test get tenant devices
test_tenant_devices() {
    if [ -z "$TENANT_ID" ]; then
        print_msg "$RED" "Error: -t <tenant_id> is required"
        exit 1
    fi

    local url="${SERVER}/api/v1/tenants/${TENANT_ID}/devices"
    print_request "GET" "$url"

    local response=$(curl -s -w "\n%{http_code}" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    print_response "$http_code" "$body"
}

# Parse command line arguments
while getopts "s:vi:u:k:p:c:d:f:t:h" opt; do
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
        u)
            DEVICE_UID="$OPTARG"
            ;;
        k)
            KEYS="$OPTARG"
            ;;
        p)
            PROPERTIES_JSON="$OPTARG"
            ;;
        c)
            CONTENT="$OPTARG"
            ;;
        d)
            DURATION="$OPTARG"
            ;;
        f)
            DEV_CODE="$OPTARG"
            ;;
        t)
            TENANT_ID="$OPTARG"
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

print_msg "$BLUE" "Testing wisefido-qinglan API"
print_msg "$BLUE" "Server: $SERVER"
print_msg "$BLUE" "Interface: $INTERFACE"
if [ -n "$DEVICE_UID" ]; then
    print_msg "$BLUE" "Device UID: $DEVICE_UID"
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
    info)
        test_info
        ;;
    properties)
        test_properties
        ;;
    properties-set)
        test_properties_set
        ;;
    subscribe)
        test_subscribe
        ;;
    function)
        test_function
        ;;
    tenant-devices)
        test_tenant_devices
        ;;
    *)
        print_msg "$RED" "Error: Unknown interface '$INTERFACE'"
        echo ""
        usage
        ;;
esac

echo ""
