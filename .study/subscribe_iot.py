#!/usr/bin/env python3
"""
Subscribe to iot:* Redis Streams directly (bypass card aggregation) and print raw JSON.

Streams:
  - iot:monitor:stream  (real-time vitals/tracks, 30s TTL)
  - iot:event:stream    (device events: InBed/LeftBed/Fall/activity, 24h TTL)
  - iot:alarm:stream    (device alarms: HeartRate/RespRate/Fall, 24h TTL)
  - iot:stat:stream     (periodic statistics, 5min TTL)
  - iot:auth:stream     (device auth requests/responses, 24h TTL)

Each message carries device_uid, device_type, tenant_id, timestamp, topic_type,
category, dataValue -- no card_id dependency.

Usage:
  pip install redis
  python .study/subscribe_iot.py
  python .study/subscribe_iot.py --streams monitor,alarm
  python .study/subscribe_iot.py --device-uid BM87XXXX
"""

import argparse
import json
import signal
import sys
import time

import redis

REDIS_HOST = "127.0.0.1"
REDIS_PORT = 6379
REDIS_PASSWORD = "TeLunSu-36kr"
REDIS_DB = 0

ALL_STREAMS = [
    "iot:monitor:stream",
    "iot:event:stream",
    "iot:alarm:stream",
    "iot:stat:stream",
    "iot:auth:stream",
]

CONSUMER_GROUP = "study-iot-group"
CONSUMER_NAME = "study-iot-consumer-1"

READ_COUNT = 20
READ_BLOCK_MS = 2000

running = True


def signal_handler(sig, frame):
    global running
    print("\n[exit] shutting down...")
    running = False


def parse_args():
    parser = argparse.ArgumentParser(description="Subscribe to iot:* Redis Streams")
    parser.add_argument(
        "--streams",
        type=str,
        default=None,
        help="Comma-separated stream short names to subscribe (monitor,event,alarm,stat,auth). Default: all",
    )
    parser.add_argument(
        "--device-uid",
        type=str,
        default=None,
        help="Filter messages by device_uid (optional)",
    )
    return parser.parse_args()


def resolve_streams(streams_arg):
    if streams_arg is None:
        return ALL_STREAMS
    names = [s.strip() for s in streams_arg.split(",")]
    result = []
    for name in names:
        full = f"iot:{name}:stream"
        if full in ALL_STREAMS:
            result.append(full)
        else:
            print(f"[warn] unknown stream '{name}', skipping. Available: monitor,event,alarm,stat,auth")
    if not result:
        print("[error] no valid streams selected", file=sys.stderr)
        sys.exit(1)
    return result


def ensure_consumer_groups(r, streams):
    for stream in streams:
        try:
            r.xgroup_create(stream, CONSUMER_GROUP, id="0", mkstream=True)
            print(f"[init] created consumer group '{CONSUMER_GROUP}' on '{stream}'")
        except redis.ResponseError as e:
            if "BUSYGROUP" in str(e):
                print(f"[init] consumer group '{CONSUMER_GROUP}' already exists on '{stream}'")
            else:
                raise


def print_message(stream_name, msg_id, fields):
    output = {
        "stream": stream_name,
        "msg_id": msg_id,
        "fields": fields,
    }
    print(json.dumps(output, ensure_ascii=False, indent=2))


def main():
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    args = parse_args()
    streams = resolve_streams(args.streams)
    device_filter = args.device_uid

    r = redis.Redis(
        host=REDIS_HOST,
        port=REDIS_PORT,
        password=REDIS_PASSWORD,
        db=REDIS_DB,
        decode_responses=True,
    )

    try:
        r.ping()
        print(f"[init] connected to Redis {REDIS_HOST}:{REDIS_PORT}")
    except redis.ConnectionError as e:
        print(f"[error] cannot connect to Redis: {e}", file=sys.stderr)
        sys.exit(1)

    ensure_consumer_groups(r, streams)

    print(f"[init] subscribing to: {', '.join(streams)}")
    if device_filter:
        print(f"[init] filtering by device_uid: {device_filter}")
    print(f"[init] consumer group={CONSUMER_GROUP}  consumer={CONSUMER_NAME}")
    print(f"[init] waiting for messages... (Ctrl+C to stop)\n")

    msg_count = 0
    stream_keys = {s: ">" for s in streams}

    while running:
        try:
            results = r.xreadgroup(
                CONSUMER_GROUP,
                CONSUMER_NAME,
                stream_keys,
                count=READ_COUNT,
                block=READ_BLOCK_MS,
            )
        except redis.ConnectionError:
            print("[warn] redis connection lost, reconnecting in 3s...")
            time.sleep(3)
            continue

        if not results:
            continue

        for stream_name, messages in results:
            for msg_id, fields in messages:
                # Filter by device_uid if specified
                if device_filter:
                    uid = fields.get("device_uid", "")
                    if uid != device_filter:
                        r.xack(stream_name, CONSUMER_GROUP, msg_id)
                        continue

                msg_count += 1
                print_message(stream_name, msg_id, fields)
                r.xack(stream_name, CONSUMER_GROUP, msg_id)

    print(f"\n[exit] total messages received: {msg_count}")
    r.close()


if __name__ == "__main__":
    main()
