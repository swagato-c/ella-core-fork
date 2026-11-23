// SPDX-FileCopyrightText: Ella Networks Inc.
// SPDX-License-Identifier: BUSL-1.1

package ue

import (
	"fmt"

	"github.com/ellanetworks/core/nas/fgs"
	"github.com/ellanetworks/core/tester/logger"
	"go.uber.org/zap"
)

func handleServiceAccept(ue *UE, plain []byte) error {
	if _, err := fgs.ParseServiceAccept(plain); err != nil {
		return fmt.Errorf("could not parse Service Accept: %v", err)
	}

	logger.UeLogger.Debug("Received Service Accept NAS message", zap.String("IMSI", ue.UeSecurity.Supi))

	return nil
}
