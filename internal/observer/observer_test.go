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

func TestFreshness(t *testing.T) {
	now := time.Now()
	deletedAt_deleted_node := now.Add(30 * time.Minute)
	testCases := []struct {
		name              string
		nodeLogMap        observer.NodeLogMapType
		visitHistory      observer.VisitHistoryType
		ip                observer.IPAddr
		at                time.Time
		expectedFreshness float64
	}{
		{
			name: "zeroVisit",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
			visitHistory:      observer.VisitHistoryType{},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 0,
		},
		{
			name: "different ip visit",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: &deletedAt_deleted_node,
				},
			},
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.2",
					NodeID:     "node1",
					VisitedAt:  now,
				},
			},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 0,
		},
		{
			name: "deleted node",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: &deletedAt_deleted_node,
				},
			},
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
			},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 0,
		},
		{
			name: "simple visit",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
			},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 1,
		},
		{
			name: "repeated visit",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now.Add(15 * time.Minute),
				},
			},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 1,
		},
		{
			name: "multiple node visit",
			nodeLogMap: observer.NodeLogMapType{
				"node1": observer.NodeLog{
					ID:        "node1",
					CreatedAt: now,
					DeletedAt: nil,
				},
				"node2": observer.NodeLog{
					ID:        "node2",
					CreatedAt: now,
					DeletedAt: nil,
				},
				"node3": observer.NodeLog{
					ID:        "node3",
					CreatedAt: now,
					DeletedAt: nil,
				},
				"node4": observer.NodeLog{
					ID:        "node4",
					CreatedAt: now,
					DeletedAt: nil,
				},
			},
			visitHistory: observer.VisitHistoryType{
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node1",
					VisitedAt:  now,
				},
				observer.VisitLog{
					RemoteAddr: "1.1.1.1",
					NodeID:     "node4",
					VisitedAt:  now,
				},
				observer.VisitLog{
					RemoteAddr: "1.1.1.2",
					NodeID:     "node3",
					VisitedAt:  now,
				},
			},
			ip:                "1.1.1.1",
			at:                now.Add(1 * time.Hour),
			expectedFreshness: 0.5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := observer.New()

			for _, nodeLog := range tc.nodeLogMap {
				o.LogNode(nodeLog)
			}

			for _, visitLog := range tc.visitHistory {
				o.LogVisit(visitLog)
			}

			assert.Equal(t, tc.expectedFreshness, o.GetFreshness(string(tc.ip), tc.at))
		})
	}
}
