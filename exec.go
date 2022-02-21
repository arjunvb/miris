package main

import (
	"./data"
	"./exec"
	"./miris"

	"fmt"
	"log"
	"os"
)

func main() {
	predName := os.Args[1]
	planFname := os.Args[2]
	sourceDir := os.Args[3]

	ppCfg, modelCfg := data.Get(predName)
	detectionPath, framePath := sourceDir+"0-detections.json", "/data/miris/data/exp/frames/0/"
	// detectionPath = "data/exp/json/0-detections.json"
	// framePath = "data/exp/frames/0/"
	var plan miris.PlannerConfig
	miris.ReadJSON(planFname, &plan)
	execCfg := miris.ExecConfig{
		DetectionPath:     detectionPath,
		FramePath:         framePath,
		TrackOutput:       fmt.Sprintf("logs/%s/%d/%v/track.json", predName, plan.Freq, plan.Bound),
		FilterOutput:      fmt.Sprintf("logs/%s/%d/%v/filter.json", predName, plan.Freq, plan.Bound),
		UncertaintyOutput: fmt.Sprintf("logs/%s/%d/%v/uncertainty.json", predName, plan.Freq, plan.Bound),
		RefineOutput:      fmt.Sprintf("logs/%s/%d/%v/refine.json", predName, plan.Freq, plan.Bound),
		OutPath:           fmt.Sprintf("%s/final.json", sourceDir), // modify  here
	}
	log.Printf("%v", modelCfg)
	exec.Exec(ppCfg, modelCfg, plan, execCfg)
}
