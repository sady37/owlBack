#!/usr/bin/env python3
"""
Subscribe to wisefido-cardagg output streams and print raw JSON messages.

Streams:
  - card:realtime:stream  (real-time vitals/tracks per device, ~1s interval)
  - card:status:stream    (state change events: alarm, bed, device status, etc.)

Usage:
  pip install redis
  python .study/subscribe_cardagg.py
"""

import json
import signal
import sys
import time

import redis

REDIS_HOST = "127.0.0.1"
REDIS_PORT = 6379
REDIS_PASSWORD = "TeLunSu-36kr"
REDIS_DB = 0

STREAMS = {
    "card:realtime:stream": ">",
    "card:status:stream": ">",
}
CONSUMER_GROUP = "study-test-group"
CONSUMER_NAME = "study-consumer-1"

READ_COUNT = 20
READ_BLOCK_MS = 2000

running = True


def signal_handler(sig, frame):
    global running
    print("\n[exit] shutting down...")
    running = False


def ensure_consumer_groups(r: redis.Redis):
    for stream in STREAMS:
        try:
            r.xgroup_create(stream, CONSUMER_GROUP, id="0", mkstream=True)
            print(f"[init] created consumer group '{CONSUMER_GROUP}' on '{stream}'")
        except redis.ResponseError as e:
            if "BUSYGROUP" in str(e):
                print(f"[init] consumer group '{CONSUMER_GROUP}' already exists on '{stream}'")
            else:
                raise


def print_message(stream_name: str, msg_id: str, fields: dict):
    """Print raw stream message as JSON."""
    output = {
        "stream": stream_name,
        "msg_id": msg_id,
        "fields": fields,
    }
    print(json.dumps(output, ensure_ascii=False, indent=2))


def main():
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

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

    ensure_consumer_groups(r)

    print(f"[init] subscribing to: {', '.join(STREAMS.keys())}")
    print(f"[init] consumer group={CONSUMER_GROUP}  consumer={CONSUMER_NAME}")
    print(f"[init] waiting for messages... (Ctrl+C to stop)\n")

    msg_count = 0
    stream_keys = {s: ">" for s in STREAMS}

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
                msg_count += 1
                print_message(stream_name, msg_id, fields)
                r.xack(stream_name, CONSUMER_GROUP, msg_id)

    print(f"\n[exit] total messages received: {msg_count}")
    r.close()


if __name__ == "__main__":
    main()
