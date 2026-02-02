package itemtransfer

import (
	"encoding/json"
	"fmt"

	"github.com/MaaXYZ/maa-framework-go/v3"
	"github.com/rs/zerolog/log"
)

func runLocate(ctx *maa.Context, arg *maa.CustomRecognitionArg, targetInv Inventory, currentNodeName string) (*maa.CustomRecognitionResult, bool) {
	var taskParam map[string]any

	err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &taskParam)
	if err != nil {
		log.Error().
			Err(err).
			Str("raw_json", arg.CustomRecognitionParam).
			Msg("Seems that we have received bad params")
		return nil, false
	}

	itemName, ok := taskParam["ItemName"].(string)
	if !ok {
		log.Error().
			Str("raw_json", arg.CustomRecognitionParam).
			Msg("ItemName is not a string")
		return nil, false
	}
	category, _ := taskParam["Category"].(string)
	//containerContent := userSetting["ContainerContent"] //todo put this into use
	var taskName string

	// 简单的映射逻辑
	switch category {
	case "Material":
		taskName = "ItemTransferSwitchToMaterial"
	case "Plant":
		taskName = "ItemTransferSwitchToPlant"
	case "Product":
		taskName = "ItemTransferSwitchToProduct"
		// case "All": ...
	}
	if taskName != "" && targetInv == REPOSITORY {
		// 🔥 直接调用 Pipeline 节点！
		// 这是一个同步调用，会等点击完成、post_wait 结束后才返回
		status := ctx.RunTask(taskName).Status

		if !status.Success() {
			log.Warn().Str("task", taskName).Msg("Failed to switch category tab, trying scan anyway...")
			// 这里可以选择 return nil, false 报错，也可以硬着头皮继续扫（万一已经在那页了呢）
		} else {
			log.Debug().Msg("Category switch successful.")
		}
	}

	log.Debug().
		Str("ItemName", itemName).
		Str("Target", targetInv.String()).
		Any("ContainerContent", taskParam["ContainerContent"]).
		Msg("Task parameters initialized")

	maxCols := targetInv.Columns()
	maxRows := RowsPerPage // 4行
	for row := range maxRows {
		for col := range maxCols {

			// Step 1 & 2
			img := MoveAndShot(ctx, targetInv, row, col)
			if img == nil {
				continue
			}
			// Step 3 - Call original OCR
			log.Debug().Msg("Starting Recognition")
			detail := ctx.RunRecognitionDirect(
				maa.NodeRecognitionTypeOCR,
				maa.NodeOCRParam{
					ROI: maa.NewTargetRect(
						TooltipRoi(targetInv, row, col),
					),
					OrderBy:  "Expected",
					Expected: []string{itemName},
				},
				img,
			)
			log.Debug().Msg("Done Recognition!!!!!")
			log.Debug().Str("detail_json", detail.DetailJson).Msg("Item OCR Full Detail")
			if detail.Hit {
				log.Info().
					Int("grid_row_y", row).
					Int("grid_col_x", col).
					Msg("Yes That's it! We have found proper item.")

				// saving cache todo move standalone
				template := "{\"ItemTransferToBackpack\": {\"recognition\": {\"param\": {\"custom_recognition_param\": {\"ItemLastFoundRowAbs\": %d,\"ItemLastFoundColumnX\": %d,\"FirstRun\": false}}}}}"
				defer ctx.OverridePipeline(fmt.Sprintf(template, row, col))

				return &maa.CustomRecognitionResult{
					Box:    ItemBoxRoi(targetInv, row, col),
					Detail: detail.DetailJson,
				}, true
			} else {
				log.Info().
					Int("grid_row_y", row).
					Int("grid_col_x", col).
					Msg("Not this one. Bypass.")
			}

		}

	}
	log.Warn().
		Msg("No item with given name found. Please check input")
	return nil, false
	//todo: switch to next page

}

// const (
// 	OCRFilter = "^(?![^a-zA-Z0-9]*(?:升序|降序|默认|品质|一键存放|材料|战术物品|消耗品|功能设备|普通设备|培养晶核)[^a-zA-Z0-9]*$)[^a-zA-Z0-9]+$"
// )

type RepoLocate struct{}

func (*RepoLocate) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	// 强制指定 REPOSITORY
	// 强制指定节点名 ItemTransferToBackpack 用于缓存
	return runLocate(ctx, arg, REPOSITORY, "ItemTransferToBackpack")
}

type BackpackLocate struct{}

func (*BackpackLocate) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	// 强制指定 BACKPACK
	// 强制指定节点名 ItemTransferToRepository 用于缓存
	return runLocate(ctx, arg, BACKPACK, "ItemTransferToRepository")
}
