#!/usr/bin/env python3
"""Compute deterministic integration-lane scorecard and selection decision."""

from __future__ import annotations

import argparse
import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

DEFAULT_WEIGHTS = {
    "setup_time": 0.25,
    "failure_rate": 0.25,
    "determinism": 0.2,
    "policy_correctness": 0.2,
    "distribution_reach": 0.1,
}

MAX_SETUP_MINUTES = 30.0


@dataclass(frozen=True)
class LaneInput:
    lane_id: str
    lane_name: str
    setup_minutes_p50: float
    failure_rate: float
    determinism_pass_rate: float
    policy_correctness_rate: float
    distribution_reach: float


def fail(message: str) -> None:
    raise SystemExit(f"integration lane scorecard failed: {message}")


def require_str(payload: dict[str, Any], field: str) -> str:
    value = payload.get(field)
    if not isinstance(value, str) or not value.strip():
        fail(f"{field} must be a non-empty string")
    return value.strip()


def require_probability(payload: dict[str, Any], field: str) -> float:
    value = payload.get(field)
    if not isinstance(value, (int, float)):
        fail(f"{field} must be a number in [0,1]")
    as_float = float(value)
    if as_float < 0.0 or as_float > 1.0:
        fail(f"{field} must be in [0,1], got {as_float}")
    return as_float


def require_non_negative(payload: dict[str, Any], field: str) -> float:
    value = payload.get(field)
    if not isinstance(value, (int, float)):
        fail(f"{field} must be a non-negative number")
    as_float = float(value)
    if as_float < 0.0:
        fail(f"{field} must be >= 0, got {as_float}")
    return as_float


def parse_lane(raw: Any, index: int) -> LaneInput:
    if not isinstance(raw, dict):
        fail(f"lanes[{index}] must be an object")
    lane_id = require_str(raw, "lane_id")
    lane_name = require_str(raw, "lane_name")
    return LaneInput(
        lane_id=lane_id,
        lane_name=lane_name,
        setup_minutes_p50=require_non_negative(raw, "setup_minutes_p50"),
        failure_rate=require_probability(raw, "failure_rate"),
        determinism_pass_rate=require_probability(raw, "determinism_pass_rate"),
        policy_correctness_rate=require_probability(raw, "policy_correctness_rate"),
        distribution_reach=require_probability(raw, "distribution_reach"),
    )


def load_input(path: Path) -> tuple[str, str, str, list[LaneInput]]:
    if not path.exists():
        fail(f"input file not found: {path}")
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as err:
        fail(f"input is not valid JSON: {err}")
    if not isinstance(payload, dict):
        fail("input root must be an object")

    schema_id = require_str(payload, "schema_id")
    schema_version = require_str(payload, "schema_version")
    source_created_at = str(payload.get("created_at", "")).strip()
    if source_created_at == "":
        source_created_at = "1980-01-01T00:00:00Z"
    raw_lanes = payload.get("lanes")
    if not isinstance(raw_lanes, list) or not raw_lanes:
        fail("lanes must be a non-empty array")

    lanes = [parse_lane(raw, idx) for idx, raw in enumerate(raw_lanes)]
    lane_ids = [lane.lane_id for lane in lanes]
    if len(lane_ids) != len(set(lane_ids)):
        fail("lane_id values must be unique")
    return schema_id, schema_version, source_created_at, lanes


def setup_score(setup_minutes_p50: float) -> float:
    if setup_minutes_p50 <= 0.0:
        return 1.0
    capped = min(setup_minutes_p50, MAX_SETUP_MINUTES)
    return max(0.0, 1.0 - (capped / MAX_SETUP_MINUTES))


def round6(value: float) -> float:
    return round(value, 6)


def evaluate_lane(lane: LaneInput) -> dict[str, Any]:
    metric_scores = {
        "setup_time": setup_score(lane.setup_minutes_p50),
        "failure_rate": max(0.0, 1.0 - lane.failure_rate),
        "determinism": lane.determinism_pass_rate,
        "policy_correctness": lane.policy_correctness_rate,
        "distribution_reach": lane.distribution_reach,
    }
    weighted = 0.0
    for key, weight in DEFAULT_WEIGHTS.items():
        weighted += metric_scores[key] * weight

    return {
        "lane_id": lane.lane_id,
        "lane_name": lane.lane_name,
        "metrics": {
            "setup_minutes_p50": round6(lane.setup_minutes_p50),
            "failure_rate": round6(lane.failure_rate),
            "determinism_pass_rate": round6(lane.determinism_pass_rate),
            "policy_correctness_rate": round6(lane.policy_correctness_rate),
            "distribution_reach": round6(lane.distribution_reach),
        },
        "normalized_scores": {key: round6(value) for key, value in sorted(metric_scores.items())},
        "weighted_score": round6(weighted),
    }


def rank_lanes(lanes: list[LaneInput]) -> list[dict[str, Any]]:
    scored = [evaluate_lane(lane) for lane in lanes]
    scored.sort(key=lambda item: (-float(item["weighted_score"]), str(item["lane_id"])))
    return scored


def decision_for(scored: list[dict[str, Any]]) -> dict[str, Any]:
    selected = scored[0]
    threshold_score = 0.75
    confidence = 1.0
    runner_up = None
    if len(scored) > 1:
        runner_up = scored[1]
        confidence = max(0.0, float(selected["weighted_score"]) - float(runner_up["weighted_score"]))

    threshold_met = float(selected["weighted_score"]) >= threshold_score
    confidence_level = "high"
    if confidence < 0.03:
        confidence_level = "low"
    elif confidence < 0.08:
        confidence_level = "medium"

    return {
        "selected_lane_id": selected["lane_id"],
        "selected_lane_name": selected["lane_name"],
        "selected_weighted_score": selected["weighted_score"],
        "runner_up_lane_id": runner_up["lane_id"] if runner_up else "",
        "runner_up_weighted_score": runner_up["weighted_score"] if runner_up else 0,
        "confidence_delta": round6(confidence),
        "confidence_level": confidence_level,
        "threshold_score": threshold_score,
        "threshold_met": threshold_met,
        "expansion_allowed": threshold_met and confidence >= 0.03,
        "decision_rule": (
            "select highest weighted_score; expansion requires selected_score>=0.75 "
            "and confidence_delta>=0.03"
        ),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="Compute integration lane scorecard")
    parser.add_argument("--input", required=True, help="path to adoption metrics JSON")
    parser.add_argument("--out", required=True, help="path to scorecard output JSON")
    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.out)
    schema_id, schema_version, source_created_at, lanes = load_input(input_path)
    scored = rank_lanes(lanes)
    decision = decision_for(scored)

    output = {
        "schema_id": "gait.launch.integration_lane_scorecard",
        "schema_version": "1.0.0",
        "created_at": source_created_at,
        "source_schema_id": schema_id,
        "source_schema_version": schema_version,
        "weights": {key: round6(value) for key, value in sorted(DEFAULT_WEIGHTS.items())},
        "max_setup_minutes": MAX_SETUP_MINUTES,
        "lanes": scored,
        "decision": decision,
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(output, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(f"integration lane scorecard written: {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
