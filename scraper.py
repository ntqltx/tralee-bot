from __future__ import annotations

import json
import os
import re
import sys
import xml.etree.ElementTree as ET
from dataclasses import asdict, dataclass
from typing import Iterable

import requests

FEED_URL = (
    "https://www.daft.ie/rss.daft?"
    "uid=0&id=0&type=rental&fri_id=0&county=kerry&area=tralee"
)
LISTING_ID_RE = re.compile(r"/(\d+)(?:[/?#]|$)")
REQUEST_TIMEOUT_SECONDS = 20
USER_AGENT = (
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 "
    "(KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)

@dataclass(frozen=True)
class Listing:
    id: int
    title: str
    url: str

def log(message: str) -> None:
    print(message, flush=True)

def require_env(name: str) -> str:
    value = os.environ.get(name, "").strip().rstrip("/")
    if not value:
        raise RuntimeError(f"{name} is required")
    return value

def text_or_empty(parent: ET.Element, child_name: str) -> str:
    child = parent.find(child_name)
    if child is None or child.text is None:
        return ""
    return child.text.strip()

def extract_listing_id(url: str) -> int | None:
    matches = LISTING_ID_RE.findall(url)
    if not matches:
        return None
    return int(matches[-1])

def fetch_feed() -> bytes:
    headers = {
        "Accept": "application/rss+xml, application/xml;q=0.9, */*;q=0.8",
        "User-Agent": USER_AGENT,
    }
    log(f"Fetching RSS feed: {FEED_URL}")
    response = requests.get(FEED_URL, headers=headers, timeout=REQUEST_TIMEOUT_SECONDS)
    response.raise_for_status()

    log(f"Fetched {len(response.content)} bytes from Daft RSS")
    return response.content

def parse_listings(feed_xml: bytes) -> list[Listing]:
    try:
        root = ET.fromstring(feed_xml)
        
    except ET.ParseError as exc:
        raise RuntimeError(f"failed to parse RSS XML: {exc}") from exc

    listings: list[Listing] = []
    for item in root.findall(".//item"):
        title = text_or_empty(item, "title")
        link = text_or_empty(item, "link")

        if not title or not link:
            log("Skipping RSS item with missing title or link")
            continue

        listing_id = extract_listing_id(link)
        if listing_id is None:
            log(f"Skipping RSS item because no numeric listing ID was found in URL: {link}")
            continue

        listings.append(Listing(id=listing_id, title=title, url=link))

    log(f"Parsed {len(listings)} listing(s)")
    return listings

def post_listings(railway_bot_url: str, listings: Iterable[Listing]) -> None:
    webhook_url = f"{railway_bot_url}/webhook/new-listings"
    payload = [asdict(listing) for listing in listings]

    log(f"Posting {len(payload)} listing(s) to {webhook_url}")
    response = requests.post(
        webhook_url,
        data=json.dumps(payload),
        headers={"Content-Type": "application/json", "User-Agent": USER_AGENT},
        timeout=REQUEST_TIMEOUT_SECONDS,
    )
    response.raise_for_status()
    log(f"Webhook accepted payload: HTTP {response.status_code} {response.text.strip()}")

def main() -> int:
    try:
        railway_bot_url = require_env("RAILWAY_BOT_URL")
        feed_xml = fetch_feed()
        listings = parse_listings(feed_xml)
        post_listings(railway_bot_url, listings)

    except requests.HTTPError as exc:
        response = exc.response
        body = response.text[:500] if response is not None else ""
        log(f"HTTP error: {exc}; response body: {body}")
        return 1

    except requests.RequestException as exc:
        log(f"Network error: {exc}")
        return 1

    except Exception as exc:
        log(f"Fatal error: {exc}")
        return 1

    log("Scrape-and-push job completed successfully")
    return 0

if __name__ == "__main__":
    sys.exit(main())