package exec

import (
	"log"
	"os"

	gnnlib "../gnn"
	"../miris"
)

// Returns needed [2]int{frameIdx, freq} from a list of frames that we need
func getNeededSpecs(needed []int, seenFrames map[int]bool, maxFrame int) [][2]int {
	var frames [][2]int
	for _, frameIdx := range needed {
		if seenFrames[frameIdx] || frameIdx < 0 || frameIdx > maxFrame {
			continue
		}
		idx1 := -1
		idx2 := -1
		for seenIdx := range seenFrames {
			if seenIdx < frameIdx && (idx1 == -1 || seenIdx > idx1) {
				idx1 = seenIdx
			} else if seenIdx > frameIdx && (idx2 == -1 || seenIdx < idx2) {
				idx2 = seenIdx
			}
		}
		freq1 := frameIdx - idx1
		freq2 := idx2 - frameIdx
		frames = append(frames, [2]int{idx1, freq1})
		frames = append(frames, [2]int{frameIdx, freq2})
	}
	for _, frameSpec := range frames {
		seenFrames[frameSpec[0]] = true
		seenFrames[frameSpec[0]+frameSpec[1]] = true
	}
	return frames
}

type GraphWithSeen struct {
	Graph []gnnlib.Edge
	Seen  map[int]bool
}

func ReadGraphAndSeen(fname string) ([]gnnlib.Edge, map[int]bool) {
	var x GraphWithSeen
	miris.ReadJSON(fname, &x)
	return x.Graph, x.Seen
}

func Exec(ppCfg miris.PreprocessConfig, modelCfg miris.ModelConfig, plan miris.PlannerConfig, execCfg miris.ExecConfig) {
	var gnnPath string
	for _, gnnCfg := range modelCfg.GNN {
		log.Printf("gnnCfg Frequency: %v, plan Frequency: %v", gnnCfg.Freq, plan.Freq)
		if gnnCfg.Freq != plan.Freq {
			continue
		}
		gnnPath = gnnCfg.ModelPath
	}
	gnn := gnnlib.NewGNN(gnnPath, execCfg.DetectionPath, execCfg.FramePath, ppCfg.FrameScale)
	defer gnn.Close()
	seenFrames := make(map[int]bool)
	var graph []gnnlib.Edge
	if _, err := os.Stat(execCfg.TrackOutput); err != nil {
		log.Printf("[exec] run initial tracking")

		for _, freq := range []int{2 * plan.Freq, plan.Freq} {
			var frames [][2]int
			log.Printf("got here in exec.go")
			for frameIdx := 0; frameIdx < gnn.NumFrames()-freq; frameIdx += freq {
				frames = append(frames, [2]int{frameIdx, freq})
				seenFrames[frameIdx] = true
				seenFrames[frameIdx+freq] = true
			}
			graph = gnn.Update(graph, frames, plan.Q)
		}
		miris.WriteJSON(execCfg.TrackOutput, GraphWithSeen{graph, seenFrames})
	} else {
		log.Printf("[exec] read track output")
		graph, seenFrames = ReadGraphAndSeen(execCfg.TrackOutput)
	}
	log.Printf("[exec] ... tracking yields graph with %d edges (seen %d frames)", len(graph), len(seenFrames))
	// maxFrame := ((gnn.NumFrames() - 1) / plan.Freq) * plan.Freq

	// extract tracks
	var components [][]gnnlib.Edge
	components = gnn.GetComponents(graph)
	var tracks [][]miris.Detection
	for _, comp := range components {
		for _, track := range gnn.SampleComponent(comp) {
			log.Print("hello")
			for i := range track {
				log.Printf("[exec] Track idx: %d, length of tracks: %d", i, len(tracks))
				track[i].TrackID = len(tracks) + 1
			}
			tracks = append(tracks, track)
		}
	}
	log.Printf("[exec] extracted %d tracks", len(tracks))
	miris.WriteJSON(execCfg.OutPath, miris.TracksToDetections(tracks))
}
