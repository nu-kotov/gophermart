package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/nu-kotov/gophermart/internal/logger"
	"github.com/nu-kotov/gophermart/internal/models"
)

func (hnd *Handler) GetAccrualPoints() {
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		unprocessedOrders, err := hnd.Storage.SelectUnprocessedOrders(context.Background())
		if err != nil {
			logger.Log.Info(err.Error())
			continue
		}
		if len(unprocessedOrders) == 0 {
			continue
		}

		client := resty.New()
		for _, order := range unprocessedOrders {
			strNum := strconv.FormatInt(order.Number, 10)

			resp, err := client.R().Get(hnd.Config.AccrualAddr + "/api/orders/" + strNum)
			if err != nil {
				logger.Log.Info(err.Error())
				continue
			}
			if resp.StatusCode() == http.StatusNoContent || resp.StatusCode() == http.StatusTooManyRequests {
				continue
			}

			var accrualData models.AccrualResponse
			err = json.Unmarshal(resp.Body(), &accrualData)
			if err != nil {
				logger.Log.Info(err.Error())
				continue
			}
			if accrualData.Status == "PROCESSING" || accrualData.Status == "REGISTERED" || accrualData.Status == "PROCESSED" || accrualData.Status == "INVALID" {
				order.Accrual = accrualData.Accrual
				order.Status = accrualData.Status

				hnd.SaveAccrualPointsCh <- order
			}
		}
	}
}
