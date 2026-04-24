package services

import (
	"strconv"

	modeltypes "github.com/elliotxin/ai-hms-backend/internal/models/types"
	"github.com/elliotxin/ai-hms-backend/internal/utils/idgen"
)

func parseLegacyID(raw string) (modeltypes.LegacyID, error) {
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}

	return modeltypes.LegacyID(v), nil
}

func nextLegacyID() (modeltypes.LegacyID, error) {
	v, err := idgen.NextID()
	if err != nil {
		return 0, err
	}

	return modeltypes.LegacyID(v), nil
}

func legacyIDString(id modeltypes.LegacyID) string {
	return strconv.FormatInt(id.Int64(), 10)
}
