// flow/executor.go
package bot

import (
	"discord-bot-service/internal/models"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type NodeResult struct {
	Type     string
	Continue bool
}

type NodeProps struct {
	Node    models.Node
	Message *discordgo.MessageCreate
	Session *discordgo.Session
}

// NodeExecutor 各ノードタイプの実行ロジックを定義する関数型
type NodeExecutor func(NodeProps) (NodeResult, error)

// FlowExecutor フロー全体の実行を管理する構造体
type FlowExecutor struct {
	nodeExecutors map[string]NodeExecutor
}

// NewFlowExecutor 新しいFlowExecutorインスタンスを作成
func NewFlowExecutor() *FlowExecutor {
	return &FlowExecutor{
		nodeExecutors: make(map[string]NodeExecutor),
	}
}

// RegisterNodeExecutor 特定のノードタイプに対する実行関数を登録
func (fe *FlowExecutor) RegisterNodeExecutor(nodeType string, executor NodeExecutor) {
	fe.nodeExecutors[nodeType] = executor
}

// ExecuteFlow フロー全体を実行し、各ノードの結果を返します
func (fe *FlowExecutor) ExecuteFlow(flow models.FlowData, m *discordgo.MessageCreate, s *discordgo.Session) (map[string]NodeResult, error) {
	results := make(map[string]NodeResult)
	visited := make(map[string]bool)

	// スタートノードを探す
	startNode, err := fe.findStartNode(flow.Nodes)
	if err != nil {
		return nil, err
	}

	// スタートノードから実行を開始
	err = fe.executeNode(startNode, flow, visited, results, m, s)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// findStartNode はフロー内のスタートノードを探します
func (fe *FlowExecutor) findStartNode(nodes []models.Node) (models.Node, error) {
	for _, node := range nodes {
		if node.Type == "start" {
			return node, nil
		}
	}
	return models.Node{}, errors.New("スタートノードが見つかりません")
}

// executeNode は単一のノードを実行し、次のノードへ進みます
func (fe *FlowExecutor) executeNode(node models.Node, flow models.FlowData, visited map[string]bool, results map[string]NodeResult, m *discordgo.MessageCreate, s *discordgo.Session) error {
	// ノードが既に訪問済みの場合はスキップ（循環参照対策）
	if visited[node.ID] {
		return nil
	}
	visited[node.ID] = true

	// ノードタイプに対応する実行関数を取得
	executor, ok := fe.nodeExecutors[node.Type]
	if !ok {
		return fmt.Errorf("ノードタイプ %s に対応する実行関数が見つかりません", node.Type)
	}

	// ノードを実行
	result, err := executor(NodeProps{
		Node:    node,
		Message: m,
		Session: s,
	})
	if err != nil {
		return err
	}
	results[node.ID] = result

	if !result.Continue {
		return nil
	}

	// 次のノードを探して実行

	nextNodes := fe.findNextNodes(node.ID, flow.Edges, flow.Nodes)
	for _, nextNode := range nextNodes {
		err := fe.executeNode(nextNode, flow, visited, results, m, s)
		if err != nil {
			return err
		}
	}

	return nil
}

// findNextNodes は現在のノードから接続されている次のノードを探します
func (fe *FlowExecutor) findNextNodes(nodeID string, edges []models.Edge, nodes []models.Node) []models.Node {
	var nextNodes []models.Node
	for _, edge := range edges {
		if edge.Source == nodeID {
			for _, node := range nodes {
				if node.ID == edge.Target {
					nextNodes = append(nextNodes, node)
					break
				}
			}
		}
	}
	return nextNodes
}
