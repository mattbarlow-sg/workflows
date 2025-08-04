#!/usr/bin/env python3
"""
Simple BPMN JSON structure validator
Tests basic structure without full schema validation
"""

import json
import sys
from pathlib import Path

def check_bpmn_file(filepath):
    """Check basic BPMN file structure"""
    errors = []
    
    try:
        with open(filepath, 'r') as f:
            data = json.load(f)
    except Exception as e:
        return [f"Failed to load JSON: {e}"]
    
    # Check basic structure
    if 'process' not in data:
        errors.append("Missing 'process' root element")
        return errors
    
    process = data['process']
    
    # Check required process fields
    if 'id' not in process:
        errors.append("Process missing 'id' field")
    if 'name' not in process:
        errors.append("Process missing 'name' field")
    if 'elements' not in process:
        errors.append("Process missing 'elements' field")
    else:
        elements = process['elements']
        
        # Collect all element IDs
        all_ids = set()
        
        # Check events
        if 'events' in elements:
            for event in elements['events']:
                if 'id' not in event:
                    errors.append(f"Event missing 'id': {event}")
                else:
                    all_ids.add(event['id'])
                if 'type' not in event:
                    errors.append(f"Event {event.get('id', '?')} missing 'type'")
                
                # Check boundary events
                if event.get('type') == 'boundaryEvent':
                    if 'attachedToRef' not in event:
                        errors.append(f"Boundary event {event.get('id', '?')} missing 'attachedToRef'")
        
        # Check activities
        if 'activities' in elements:
            for activity in elements['activities']:
                if 'id' not in activity:
                    errors.append(f"Activity missing 'id': {activity}")
                else:
                    all_ids.add(activity['id'])
                if 'name' not in activity:
                    errors.append(f"Activity {activity.get('id', '?')} missing 'name'")
                if 'type' not in activity:
                    errors.append(f"Activity {activity.get('id', '?')} missing 'type'")
                
                # Check agent assignment
                if 'agent' in activity:
                    agent = activity['agent']
                    if 'type' not in agent:
                        errors.append(f"Activity {activity.get('id', '?')} agent missing 'type'")
                    if 'strategy' not in agent:
                        errors.append(f"Activity {activity.get('id', '?')} agent missing 'strategy'")
        
        # Check gateways
        if 'gateways' in elements:
            for gateway in elements['gateways']:
                if 'id' not in gateway:
                    errors.append(f"Gateway missing 'id': {gateway}")
                else:
                    all_ids.add(gateway['id'])
                if 'type' not in gateway:
                    errors.append(f"Gateway {gateway.get('id', '?')} missing 'type'")
        
        # Check sequence flows
        if 'sequenceFlows' in elements:
            for flow in elements['sequenceFlows']:
                if 'id' not in flow:
                    errors.append(f"Sequence flow missing 'id': {flow}")
                if 'sourceRef' not in flow:
                    errors.append(f"Sequence flow {flow.get('id', '?')} missing 'sourceRef'")
                elif flow['sourceRef'] not in all_ids:
                    errors.append(f"Sequence flow {flow.get('id', '?')} has invalid sourceRef: {flow['sourceRef']}")
                if 'targetRef' not in flow:
                    errors.append(f"Sequence flow {flow.get('id', '?')} missing 'targetRef'")
                elif flow['targetRef'] not in all_ids:
                    errors.append(f"Sequence flow {flow.get('id', '?')} has invalid targetRef: {flow['targetRef']}")
    
    return errors

def main():
    test_files = [
        "test-data/simple-process.json",
        "test-data/parallel-gateway.json",
        "test-data/exclusive-gateway.json",
        "test-data/subprocess.json",
        "test-data/ai-human-collab.json",
        "test-data/dynamic-assignment.json",
        "test-data/invalid-bpmn.json"
    ]
    
    print("BPMN JSON Structure Validation")
    print("=" * 50)
    
    for filepath in test_files:
        print(f"\nTesting {filepath}...")
        errors = check_bpmn_file(filepath)
        
        if not errors:
            print(f"✓ {filepath} - VALID")
        else:
            print(f"✗ {filepath} - INVALID")
            for error in errors:
                print(f"  - {error}")

if __name__ == "__main__":
    main()