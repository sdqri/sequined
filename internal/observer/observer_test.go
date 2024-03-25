package observer_test

import (
	"testing"
	"time"

	"github.com/sdqri/sequined/internal/observer"
	"github.com/stretchr/testify/assert"
)

func TestLogNode(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name               string
		nodeLogMap         observer.NodeLogMapType
		expectedNodeLogMap observer.NodeLogMapType
	}{
		{
			name: "Add one node log",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
			expectedNodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := observer.New()
			for _, nodeLog := range tc.nodeLogMap {
				o.LogNode(nodeLog)
			}

			assert.Equal(t, len(tc.expectedNodeLogMap), len(o.NodeLogMap))
			assert.EqualValues(t, tc.expectedNodeLogMap, o.NodeLogMap, "NodeHistory does not match expected")
		})
	}
}

func TestLogVisit(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name                 string
		visitHistory         observer.VisitHistoryType
		expectedVisitHistory observer.VisitHistoryType
	}{
		{
			name: "Add one visit log",
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
			},
			expectedVisitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := observer.New()
			for _, visitLog := range tc.visitHistory {
				o.LogVisit(visitLog)
			}

			assert.Equal(t, len(tc.expectedVisitHistory), len(o.VisitHistory))
			assert.EqualValues(t, tc.expectedVisitHistory, o.VisitHistory, "NodeHistory does not match expected")
		})
	}
}
