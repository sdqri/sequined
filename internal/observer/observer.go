package observer

import (
	"time"
)

type IPAddr string
type NodeID string

type VisitLog struct {
	RemoteAddr IPAddr
	NodeID     NodeID
	VisitedAt  time.Time
}

type NodeLog struct {
	ID        NodeID
	CreatedAt time.Time
	DeletedAt *time.Time
}

type NodeLogMapType map[NodeID]NodeLog
type VisitHistoryType []VisitLog

type Observer struct {
	NodeLogMap   NodeLogMapType
	VisitHistory VisitHistoryType
}

func New() *Observer {
	return &Observer{
		NodeLogMap:   make(NodeLogMapType),
		VisitHistory: make(VisitHistoryType, 0),
	}
}

func (observer *Observer) LogNode(nodeLog NodeLog) {
	observer.NodeLogMap[nodeLog.ID] = nodeLog
}

func (observer *Observer) LogVisit(visitLog VisitLog) {
	observer.VisitHistory = append(observer.VisitHistory, visitLog)
}

func (observer *Observer) GetFreshness(ip string, at time.Time) float64 {
	archiveNodesMap := make(NodeLogMapType)
	for ID, nodeLog := range observer.NodeLogMap {
		if nodeLog.CreatedAt.Before(at) && (nodeLog.DeletedAt == nil ||
			nodeLog.DeletedAt.After(at)) {
			archiveNodesMap[ID] = nodeLog
		}
	}

	visitedNodes := make(NodeLogMapType)
	for _, visitLog := range observer.VisitHistory {
		if visitLog.RemoteAddr == IPAddr(ip) && visitLog.VisitedAt.Before(at) {
			if nodeLog, ok := archiveNodesMap[visitLog.NodeID]; ok {
				visitedNodes[nodeLog.ID] = nodeLog
			}
		}
	}

	if len(archiveNodesMap) == 0 || len(visitedNodes) == 0 {
		return 0
	}

	return float64(len(visitedNodes)) / float64(len(archiveNodesMap))
}

func (observer *Observer) GetAge(ip string, at time.Time) time.Duration {
	archiveNodesMap := make(NodeLogMapType)
	for ID, nodeLog := range observer.NodeLogMap {
		if nodeLog.CreatedAt.Before(at) && (nodeLog.DeletedAt == nil ||
			nodeLog.DeletedAt.After(at)) {
			archiveNodesMap[ID] = nodeLog
		}
	}

	visitByNodeIDMap := make(map[NodeID]VisitLog)
	for _, visitLog := range observer.VisitHistory {
		if visitLog.RemoteAddr == IPAddr(ip) && visitLog.VisitedAt.Before(at) {
			visitByNodeIDMap[visitLog.NodeID] = visitLog
		}
	}

	// nodeCount := 0
	cumulativeTime := time.Duration(0)
	for nodeID, visitLog := range visitByNodeIDMap {
		if nodeLog, ok := archiveNodesMap[nodeID]; ok {
			cumulativeTime += visitLog.VisitedAt.Sub(nodeLog.CreatedAt)
		}
	}

	if len(visitByNodeIDMap) == 0 {
		return 0
	}

	return cumulativeTime / time.Duration(len(visitByNodeIDMap))
}
