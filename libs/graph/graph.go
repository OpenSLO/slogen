package graph

import (
	"github.com/AbirHamzi/dd-slogen/libs"
	"github.com/goccy/go-graphviz/cgraph"

	"github.com/goccy/go-graphviz"
)

func MakeOSLODepGraph(confs map[string]*libs.SLOMultiVerse, c libs.GenConf) *graphviz.Graphviz {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		libs.Log().Fatal(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			libs.Log().Fatal(err)
		}
		g.Close()
	}()

	err = createNodes(graph, confs)
	if err != nil {
		libs.Log().Errorw("error generating slo graph nodes", "err", err)
	}

	err = createEdges(graph, confs)
	if err != nil {
		libs.Log().Errorw("error generating slo graph edges", "err", err)
	}

	filePath := c.OutDir + "/" + "slo-dep-graph.png"

	g.RenderFilename(graph, "png", filePath)

	return g
}

func createNodes(graph *cgraph.Graph, confs map[string]*libs.SLOMultiVerse) error {

	var err error
	for _, v := range confs {
		var node *cgraph.Node
		if v.SLO != nil {
			node, err = graph.CreateNode("SLO-" + v.SLO.Metadata.Name)
			node.SetLabel(v.SLO.Metadata.Name)
			//node.SetColor("#00ff0044")
			node.SetStyle("filled")
			node.SetFillColor("#aaaaee")
			node.SetComment("SLO")
			node.SetShape("invhouse")
			node.SetTooltip("SLO")

		}

		if v.AlertPolicy != nil {
			node, err = graph.CreateNode("AP-" + v.AlertPolicy.Metadata.Name)
			node.SetLabel(v.AlertPolicy.Metadata.Name)
			node.SetStyle("filled")
			node.SetFillColor("#dd8899")
			node.SetShape("tripleoctagon")
		}

		if v.AlertNotificationTarget != nil {
			node, err = graph.CreateNode("NT-" + v.AlertNotificationTarget.Metadata.Name)
			node.SetLabel(v.AlertNotificationTarget.Metadata.Name)
			node.SetStyle("filled")
			node.SetFillColor("#ddaa77")
			node.SetShape("house")
		}

		if err != nil {
			libs.Log().Errorw("error generating slo graph", "err", err)
			return err
		}
		//node.
	}

	return nil
}

func createEdges(graph *cgraph.Graph, confs map[string]*libs.SLOMultiVerse) error {

	for _, v := range confs {

		if v.SLO != nil {
			sloNode, err := graph.Node("SLO-" + v.SLO.Metadata.Name)

			if err != nil {
				return err
			}

			apNames := v.SLO.Spec.AlertPolicies

			for _, apName := range apNames {
				apNode, err := graph.Node("AP-" + apName)
				if err != nil {
					return err
				}

				edgeName := "SLO-" + v.SLO.Metadata.Name + " -> AP-" + apName
				edge, err := graph.CreateEdge(edgeName, sloNode, apNode)
				edge.SetLabel("alert on")
				edge.SetFontSize(10)
				if err != nil {
					return err
				}
			}
		}

		if v.AlertPolicy != nil {
			apNode, err := graph.Node("AP-" + v.AlertPolicy.Metadata.Name)

			if err != nil {
				return err
			}

			ntNames := v.AlertPolicy.Spec.NotificationTargets
			for _, ntName := range ntNames {
				ntNode, err := graph.Node("NT-" + ntName.TargetRef)
				if err != nil {
					return err
				}

				edgeName := "AP-" + v.AlertPolicy.Metadata.Name + " -> NT-" + ntName.TargetRef
				edge, err := graph.CreateEdge(edgeName, apNode, ntNode)

				if err != nil {
					return err
				}
				edge.SetFontSize(10)
				edge.SetLabel("notify")
			}
		}
	}
	return nil
}
